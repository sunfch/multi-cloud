// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opensds/multi-cloud/datamover/pkg/gcp/s3"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-micro/client"
	"github.com/opensds/multi-cloud/backend/proto"
	flowtype "github.com/opensds/multi-cloud/dataflow/pkg/model"
	"github.com/opensds/multi-cloud/datamover/pkg/amazon/s3"
	"github.com/opensds/multi-cloud/datamover/pkg/azure/blob"
	"github.com/opensds/multi-cloud/datamover/pkg/ceph/s3"
	"github.com/opensds/multi-cloud/datamover/pkg/db"
	"github.com/opensds/multi-cloud/datamover/pkg/hw/obs"
	"github.com/opensds/multi-cloud/datamover/pkg/ibm/cos"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	pb "github.com/opensds/multi-cloud/datamover/proto"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
)

var simuRoutines = 10
var PART_SIZE int64 = 16 * 1024 * 1024 //The max object size that can be moved directly, default is 16M.
var JOB_RUN_TIME_MAX = 86400           //seconds, equals 1 day
var s3client osdss3.S3Service
var bkendclient backend.BackendService

var logger = log.New(os.Stdout, "", log.LstdFlags)

type Migration interface {
	Init()
	HandleMsg(msg string)
}

func Init() {
	logger.Println("Migration init.")
	s3client = osdss3.NewS3Service("s3", client.DefaultClient)
	bkendclient = backend.NewBackendService("backend", client.DefaultClient)
}

func HandleMsg(msgData []byte) error {
	var job pb.RunJobRequest
	err := json.Unmarshal(msgData, &job)
	if err != nil {
		logger.Printf("Unmarshal failed, err:%v\n", err)
		return err
	}

	//Check the status of job, and run it if needed
	status := db.DbAdapter.GetJobStatus(job.Id)
	if status != flowtype.JOB_STATUS_PENDING {
		logger.Printf("Job[ID#%s] is not in %s status.\n", job.Id, flowtype.JOB_STATUS_PENDING)
		//return errors.New("Job already running.")
		return nil //No need to consume this message again
	}

	logger.Printf("HandleMsg:job=%+v\n", job)
	go runjob(&job)
	return nil
}

func doMove(ctx context.Context, objs []*osdss3.Object, capa chan int64, th chan int, srcLoca *LocationInfo,
	destLoca *LocationInfo, remainSource bool) {
	//Only three routines allowed to be running at the same time
	//th := make(chan int, simuRoutines)
	locMap := make(map[string]*LocationInfo)
	for i := 0; i < len(objs); i++ {
		logger.Printf("************Begin to move obj(key:%s)\n", objs[i].ObjectKey)
		go move(ctx, objs[i], capa, th, srcLoca, destLoca, remainSource, locMap)
		//Create one routine
		th <- 1
		logger.Printf("doMigrate: produce 1 routine, len(th):%d.\n", len(th))
	}
}

func getOsdsLocation(ctx context.Context, virtBkname string, backendName string) (*LocationInfo, error) {
	if backendName == "" {
		logger.Println("Get backend location failed, because backend name is null.")
		return nil, errors.New("failed")
	}

	bk, err := db.DbAdapter.GetBackendByName(backendName)
	if err != nil {
		logger.Printf("Get backend information failed, err:%v\n", err)
		return nil, errors.New("failed")
	} else {
		loca := &LocationInfo{StorType:bk.Type, Region:bk.Region, EndPoint:bk.Endpoint, BucketName:bk.BucketName,
			VirBucket:virtBkname, Access:bk.Access, Security:bk.Security, BakendName:backendName}
		logger.Printf("Refresh backend[name:%s,id:%s] successfully.\n", backendName, bk.Id.String())
		return loca, nil
	}
}

func getConnLocation(ctx context.Context, conn *pb.Connector) (*LocationInfo, error) {
	switch conn.Type {
	case flowtype.STOR_TYPE_OPENSDS:
		{
			virtBkname := conn.GetBucketName()
			reqbk := osdss3.Bucket{Name: virtBkname}
			rspbk, err := s3client.GetBucket(ctx, &reqbk)
			if err != nil {
				logger.Printf("Get bucket[%s] information failed when refresh connector location, err:%v\n", virtBkname, err)
				return nil, errors.New("get bucket information failed")
			}
			return getOsdsLocation(ctx, virtBkname, rspbk.Backend)
		}
	case flowtype.STOR_TYPE_AWS_S3, flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_AZURE_BLOB, flowtype.STOR_TYPE_CEPH_S3, flowtype.STOR_TYPE_GCP_S3, flowtype.STOR_TYPE_IBM_COS:
		{
			cfg := conn.ConnConfig
			loca := LocationInfo{}
			loca.StorType = conn.Type
			for i := 0; i < len(cfg); i++ {
				switch cfg[i].Key {
				case "region":
					loca.Region = cfg[i].Value
				case "endpoint":
					loca.EndPoint = cfg[i].Value
				case "bucketname":
					loca.BucketName = cfg[i].Value
				case "access":
					loca.Access = cfg[i].Value
				case "security":
					loca.Security = cfg[i].Value
				default:
					logger.Printf("Uknow key[%s] for connector.\n", cfg[i].Key)
				}
			}
			return &loca, nil
		}
	default:
		{
			logger.Printf("Unsupport type:%s.\n", conn.Type)
			return nil, errors.New("unsupport type")
		}
	}
}

func MoveObj(obj *osdss3.Object, srcLoca *LocationInfo, destLoca *LocationInfo) error {
	logger.Printf("*****Move object[%s] from #%s# to #%s#, size is %d.\n", obj.ObjectKey, srcLoca.BakendName,
		destLoca.BakendName, obj.Size)
	if obj.Size <= 0 {
		return nil
	}
	buf := make([]byte, obj.Size)
	var size int64 = 0
	var err error = nil
	var downloader, uploader MoveWorker
	downloadObjKey := obj.ObjectKey
	if srcLoca.VirBucket != "" {
		downloadObjKey = srcLoca.VirBucket + "/" + downloadObjKey
	}
	//download
	switch srcLoca.StorType {
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		downloader = &obsmover.ObsMover{}
		size, err = downloader.DownloadObj(downloadObjKey, srcLoca, buf)
	case flowtype.STOR_TYPE_AWS_S3:
		downloader = &s3mover.S3Mover{}
		size, err = downloader.DownloadObj(downloadObjKey, srcLoca, buf)
	case flowtype.STOR_TYPE_IBM_COS:
		downloader = &ibmcosmover.IBMCOSMover{}
		size, err = downloader.DownloadObj(downloadObjKey, srcLoca, buf)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		downloader = &blobmover.BlobMover{}
		size, err = downloader.DownloadObj(downloadObjKey, srcLoca, buf)
	case flowtype.STOR_TYPE_CEPH_S3:
		downloader = &cephs3mover.CephS3Mover{}
		size, err = downloader.DownloadObj(downloadObjKey, srcLoca, buf)
	case flowtype.STOR_TYPE_GCP_S3:
		downloader = &Gcps3mover.GcpS3Mover{}
		size, err = downloader.DownloadObj(downloadObjKey, srcLoca, buf)
	default:
		{
			logger.Printf("Not support source backend type:%v\n", srcLoca.StorType)
			err = errors.New("Not support source backend type.")
		}
	}

	if err != nil {
		logger.Printf("Download object[%s] failed.", obj.ObjectKey)
		return err
	}
	logger.Printf("Download object[%s] succeed, size=%d\n", obj.ObjectKey, size)

	//upload
	uploadObjKey := obj.ObjectKey
	if srcLoca.VirBucket != "" {
		uploadObjKey = destLoca.VirBucket + "/" + uploadObjKey
	}

	switch destLoca.StorType {
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		uploader = &obsmover.ObsMover{}
		err = uploader.UploadObj(uploadObjKey, destLoca, buf)
	case flowtype.STOR_TYPE_AWS_S3:
		uploader = &s3mover.S3Mover{}
		err = uploader.UploadObj(uploadObjKey, destLoca, buf)
	case flowtype.STOR_TYPE_IBM_COS:
		uploader = &ibmcosmover.IBMCOSMover{}
		err = uploader.UploadObj(uploadObjKey, destLoca, buf)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		uploader = &blobmover.BlobMover{}
		err = uploader.UploadObj(uploadObjKey, destLoca, buf)
	case flowtype.STOR_TYPE_CEPH_S3:
		uploader = &cephs3mover.CephS3Mover{}
		err = uploader.UploadObj(uploadObjKey, destLoca, buf)
	case flowtype.STOR_TYPE_GCP_S3:
		uploader = &Gcps3mover.GcpS3Mover{}
		err = uploader.UploadObj(uploadObjKey, destLoca, buf)
	default:
		logger.Printf("Not support destination backend type:%v\n", destLoca.StorType)
		return errors.New("Not support destination backend type.")
	}
	if err != nil {
		logger.Printf("Upload object[bucket:%s,key:%s] failed, err:%v.\n", destLoca.BucketName, uploadObjKey, err)
	} else {
		logger.Printf("Upload object[bucket:%s,key:%s] successfully.\n", destLoca.BucketName, uploadObjKey)
	}

	return err
}

func multiPartDownloadInit(srcLoca *LocationInfo) (mover MoveWorker, err error) {
	switch srcLoca.StorType {
	case flowtype.STOR_TYPE_AWS_S3:
		mover := &s3mover.S3Mover{}
		err := mover.MultiPartDownloadInit(srcLoca)
		return mover, err
	case flowtype.STOR_TYPE_IBM_COS:
		mover := &ibmcosmover.IBMCOSMover{}
		err := mover.MultiPartDownloadInit(srcLoca)
		return mover, err
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		mover := &obsmover.ObsMover{}
		err := mover.MultiPartDownloadInit(srcLoca)
		return mover, err
	case flowtype.STOR_TYPE_AZURE_BLOB:
		mover := &blobmover.BlobMover{}
		err := mover.MultiPartDownloadInit(srcLoca)
		return mover, err
	case flowtype.STOR_TYPE_CEPH_S3:
		mover := &cephs3mover.CephS3Mover{}
		err := mover.MultiPartDownloadInit(srcLoca)
		return mover, err
	case flowtype.STOR_TYPE_GCP_S3:
		mover := &Gcps3mover.GcpS3Mover{}
		err := mover.MultiPartDownloadInit(srcLoca)
		return mover, err

	default:
		logger.Printf("Unsupport storType[%s] to init multipart download.\n", srcLoca.StorType)
	}

	return nil, errors.New("Unsupport storage type.")
}

func multiPartUploadInit(objKey string, destLoca *LocationInfo) (mover MoveWorker, err error) {
	switch destLoca.StorType {
	case flowtype.STOR_TYPE_AWS_S3:
		mover := &s3mover.S3Mover{}
		err := mover.MultiPartUploadInit(objKey, destLoca)
		return mover, err
	case flowtype.STOR_TYPE_IBM_COS:
		mover := &ibmcosmover.IBMCOSMover{}
		err := mover.MultiPartUploadInit(objKey, destLoca)
		return mover, err
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		mover := &obsmover.ObsMover{}
		err := mover.MultiPartUploadInit(objKey, destLoca)
		return mover, err
	case flowtype.STOR_TYPE_AZURE_BLOB:
		mover := &blobmover.BlobMover{}
		err := mover.MultiPartUploadInit(objKey, destLoca)
		return mover, err
	case flowtype.STOR_TYPE_CEPH_S3:
		mover := &cephs3mover.CephS3Mover{}
		err := mover.MultiPartUploadInit(objKey, destLoca)
		return mover, err
	case flowtype.STOR_TYPE_GCP_S3:
		mover := &Gcps3mover.GcpS3Mover{}
		err := mover.MultiPartUploadInit(objKey, destLoca)
		return mover, err
	default:
		logger.Printf("Unsupport storType[%s] to download.\n", destLoca.StorType)
	}

	return nil, errors.New("Unsupport storage type.")
}

func abortMultipartUpload(objKey string, destLoca *LocationInfo, mover MoveWorker) error {
	switch destLoca.StorType {
	case flowtype.STOR_TYPE_AWS_S3, flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE,
		flowtype.STOR_TYPE_HW_FUSIONCLOUD, flowtype.STOR_TYPE_AZURE_BLOB, flowtype.STOR_TYPE_CEPH_S3, flowtype.STOR_TYPE_GCP_S3, flowtype.STOR_TYPE_IBM_COS:
		return mover.AbortMultipartUpload(objKey, destLoca)
	default:
		logger.Printf("Unsupport storType[%s] to download.\n", destLoca.StorType)
	}

	return errors.New("Unsupport storage type.")
}

func completeMultipartUpload(objKey string, destLoca *LocationInfo, mover MoveWorker) error {
	switch destLoca.StorType {
	case flowtype.STOR_TYPE_AWS_S3, flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE,
		flowtype.STOR_TYPE_HW_FUSIONCLOUD, flowtype.STOR_TYPE_AZURE_BLOB, flowtype.STOR_TYPE_CEPH_S3, flowtype.STOR_TYPE_GCP_S3, flowtype.STOR_TYPE_IBM_COS:
		return mover.CompleteMultipartUpload(objKey, destLoca)
	default:
		logger.Printf("Unsupport storType[%s] to download.\n", destLoca.StorType)
	}

	return errors.New("Unsupport storage type.")
}

func MultipartMoveObj(obj *osdss3.Object, srcLoca *LocationInfo, destLoca *LocationInfo) error {
	partCount := int64(obj.Size / PART_SIZE)
	if obj.Size%PART_SIZE != 0 {
		partCount++
	}

	logger.Printf("*****Move object[%s] from #%s# to #%s#, size is %d.\n", obj.ObjectKey, srcLoca.BakendName,
		destLoca.BakendName, obj.Size)
	downloadObjKey := obj.ObjectKey
	if srcLoca.VirBucket != "" {
		downloadObjKey = srcLoca.VirBucket + "/" + downloadObjKey
	}
	uploadObjKey := obj.ObjectKey
	if destLoca.VirBucket != "" {
		uploadObjKey = destLoca.VirBucket + "/" + uploadObjKey
	}

	buf := make([]byte, PART_SIZE)
	var i int64
	var err error
	var uploadMover, downloadMover MoveWorker
	currPartSize := PART_SIZE
	for i = 0; i < partCount; i++ {
		partNumber := i + 1
		offset := int64(i) * PART_SIZE
		if i+1 == partCount {
			currPartSize = obj.Size - offset
			buf = nil
			buf = make([]byte, currPartSize)
		}

		//download
		start := offset
		end := offset + currPartSize - 1
		if partNumber == 1 {
			downloadMover, err = multiPartDownloadInit(srcLoca)
			if err != nil {
				return err
			}
		}
		readSize, err := downloadMover.DownloadRange(downloadObjKey, srcLoca, buf, start, end)
		if err != nil {
			return errors.New("Download failed.")
		}
		//fmt.Printf("Download part %d range[%d:%d] successfully.\n", partNumber, offset, end)
		if int64(readSize) != currPartSize {
			logger.Printf("Internal error, currPartSize=%d, readSize=%d\n", currPartSize, readSize)
			return errors.New("Internal error")
		}

		//upload
		if partNumber == 1 {
			//init multipart upload
			uploadMover, err = multiPartUploadInit(uploadObjKey, destLoca)
			if err != nil {
				return err
			}
		}
		err1 := uploadMover.UploadPart(uploadObjKey, destLoca, currPartSize, buf, partNumber, offset)
		if err1 != nil {
			err := abortMultipartUpload(obj.ObjectKey, destLoca, uploadMover)
			if err != nil {
				logger.Printf("Abort s3 multipart upload failed, err:%v\n", err)
			}
			return errors.New("S3 multipart upload failed.")
		}
		//completeParts = append(completeParts, completePart)
	}

	err = completeMultipartUpload(uploadObjKey, destLoca, uploadMover)
	if err != nil {
		logger.Println(err.Error())
		err := abortMultipartUpload(obj.ObjectKey, destLoca, uploadMover)
		if err != nil {
			logger.Printf("Abort s3 multipart upload failed, err:%v\n", err)
		}
	}

	return err
}

func deleteObj(ctx context.Context, obj *osdss3.Object, loca *LocationInfo) error {
	objKey := obj.ObjectKey
	if loca.VirBucket != "" {
		objKey = loca.VirBucket + "/" + objKey
	}
	var err error = nil
	switch loca.StorType {
	case flowtype.STOR_TYPE_AWS_S3:
		mover := s3mover.S3Mover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_IBM_COS:
		mover := ibmcosmover.IBMCOSMover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		mover := obsmover.ObsMover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		mover := blobmover.BlobMover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_CEPH_S3:
		mover := cephs3mover.CephS3Mover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_GCP_S3:
		mover := Gcps3mover.GcpS3Mover{}
		err = mover.DeleteObj(objKey, loca)
	default:
		logger.Printf("Delete object[objkey:%s] from backend storage failed.\n", obj.ObjectKey)
		err = errors.New("Unspport storage type.")
	}

	if err != nil {
		return err
	}

	//delete metadata
	if loca.VirBucket != "" {
		delMetaReq := osdss3.DeleteObjectInput{Bucket: loca.VirBucket, Key: obj.ObjectKey}
		_, err = s3client.DeleteObject(ctx, &delMetaReq)
		if err != nil {
			logger.Printf("Delete object metadata of obj[bucket:%s,objKey:%s] failed, err:%v\n", loca.VirBucket,
				obj.ObjectKey, err)
		} else {
			logger.Printf("Delete object metadata of obj[bucket:%s,objKey:%s] successfully.\n", loca.VirBucket,
				obj.ObjectKey)
		}
	}

	return err
}

func refreshSrcLocation(ctx context.Context, obj *osdss3.Object, srcLoca *LocationInfo, destLoca *LocationInfo,
	locMap map[string]*LocationInfo) (newSrcLoca *LocationInfo, err error) {
	if obj.Backend != srcLoca.BakendName && obj.Backend != "" {
		//If oject does not use the default backend
		logger.Printf("locaMap:%+v\n", locMap)
		//for selfdefined connector, obj.backend and srcLoca.backendname would be ""
		//TODO: use read/wirte lock
		newLoc, exists := locMap[obj.Backend]
		if !exists {
			newLoc, err = getOsdsLocation(ctx, obj.BucketName, obj.Backend)
			if err != nil {
				return nil, err
			}
		}
		locMap[obj.Backend] = newLoc
		logger.Printf("newSrcLoca=%+v\n", newLoc)
		return newLoc, nil
	}

	return srcLoca, nil
}

func move(ctx context.Context, obj *osdss3.Object, capa chan int64, th chan int, srcLoca *LocationInfo,
	destLoca *LocationInfo, remainSource bool, locaMap map[string]*LocationInfo) {
	logger.Printf("Obj[%s] is stored in the backend is [%s], default backend is [%s], target backend is [%s].\n",
		obj.ObjectKey, obj.Backend, srcLoca.BakendName, destLoca.BakendName)

	succeed := true
	needMove := true
	newSrcLoca, err := refreshSrcLocation(ctx, obj, srcLoca, destLoca, locaMap)
	if err != nil {
		needMove = false
		succeed = false
	}

	if needMove {
		//move object
		part_size, err := strconv.ParseInt(os.Getenv("PARTSIZE"), 10, 64)
		logger.Printf("part_size=%d, err=%v.\n", part_size, err)
		if err == nil {
			//part_size must be more than 5M and less than 100M
			if part_size >= 5 && part_size <= 100 {
				PART_SIZE = part_size * 1024 * 1024
				logger.Printf("Set PART_SIZE to be %d.\n", PART_SIZE)
			}
		}
		if obj.Size <= PART_SIZE {
			err = MoveObj(obj, newSrcLoca, destLoca)
		} else {
			err = MultipartMoveObj(obj, newSrcLoca, destLoca)
		}

		if err != nil {
			succeed = false
		}
	}

	//TODO: what if update meatadata failed
	//add object metadata to the destination bucket if destination is not self-defined
	if succeed && destLoca.VirBucket != "" {
		obj.BucketName = destLoca.VirBucket
		obj.Backend = destLoca.BakendName
		obj.LastModified = time.Now().Unix()
		_, err := s3client.CreateObject(ctx, obj)
		if err != nil {
			logger.Printf("Add object metadata of obj [objKey:%s] to bucket[name:%s] failed,err:%v.\n", obj.ObjectKey,
				obj.BucketName, err)
		} else {
			logger.Printf("Add object metadata of obj [objKey:%s] to bucket[name:%s] succeed.\n", obj.ObjectKey,
				obj.BucketName)
		}
	}

	//Delete source data if needed
	logger.Printf("remainSource for object[%s] is:%v.", obj.ObjectKey, remainSource)
	if succeed && !remainSource {
		deleteObj(ctx, obj, newSrcLoca)
		//TODO: what if delete failed
	}

	if succeed {
		//If migrate success, update capacity
		logger.Printf("  migrate object[%s] succeed.", obj.ObjectKey)
		capa <- obj.Size
	} else {
		logger.Printf("  migrate object[%s] failed.", obj.ObjectKey)
		capa <- -1
	}
	t := <-th
	logger.Printf("  migrate: consume %d routine, len(th)=%d\n", t, len(th))
}

func getOsdsS3Objs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	//TODO:need to support filter
	req := osdss3.ListObjectsRequest{Bucket: conn.BucketName}
	objs, err := s3client.ListObjects(ctx, &req)
	totalObjs := len(objs.ListObjects)
	if err != nil {
		logger.Printf("List objects failed, err:%v\n", err)
		return nil, err
	}

	srcObjs := []*osdss3.Object{}
	pattern := fmt.Sprintf("^%s", filt.Prefix)
	logger.Printf("pattern:%s\n", pattern)
	for i := 0; i < totalObjs; i++ {
		valid := true
		if filt != nil && filt.Prefix != "" {
			match, _ := regexp.MatchString(pattern, objs.ListObjects[i].ObjectKey)
			if match == false {
				valid = false
			}
		}
		if valid == true {
			srcObjs = append(srcObjs, objs.ListObjects[i])
			logger.Printf("Object[%s] match, will be migrated.\n", objs.ListObjects[i].ObjectKey)
		} else {
			logger.Printf("Object[%s] does not match, will not be migrated.\n", objs.ListObjects[i].ObjectKey)
		}
	}
	return srcObjs, nil
}

func getIBMCosObjs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	//TODO(acorbellini): reuse getAWSS3Objs function
	srcObjs := []*osdss3.Object{}
	objs, err := ibmcosmover.ListObjs(defaultSrcLoca, filt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(objs); i++ {
		obj := osdss3.Object{Size: *objs[i].Size, ObjectKey: *objs[i].Key, Backend: ""}
		srcObjs = append(srcObjs, &obj)
	}
	return srcObjs, nil
}

func getAwsS3Objs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	//TODO:need to support filter
	srcObjs := []*osdss3.Object{}
	objs, err := s3mover.ListObjs(defaultSrcLoca, filt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(objs); i++ {
		obj := osdss3.Object{Size: *objs[i].Size, ObjectKey: *objs[i].Key, Backend: ""}
		//srcObjs = append(srcObjs, &SourceOject{StorType:defaultSrcLoca.StorType, Obj:&obj})
		srcObjs = append(srcObjs, &obj)
	}
	return srcObjs, nil
}

func getHwObjs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	//TODO:need to support filter
	srcObjs := []*osdss3.Object{}
	objs, err := obsmover.ListObjs(defaultSrcLoca, filt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(objs); i++ {
		obj := osdss3.Object{Size: objs[i].Size, ObjectKey: objs[i].Key, Backend: ""}
		srcObjs = append(srcObjs, &obj)
	}
	return srcObjs, nil
}

func getAzureBlobs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	srcObjs := []*osdss3.Object{}
	objs, err := blobmover.ListObjs(defaultSrcLoca, filt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(objs); i++ {
		obj := osdss3.Object{Size: *objs[i].Properties.ContentLength, ObjectKey: objs[i].Name, Backend: ""}
		srcObjs = append(srcObjs, &obj)
	}
	return srcObjs, nil
}

//to get object details from ceph backend
func getCephS3Objs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	srcObjs := []*osdss3.Object{}
	objs, err := cephs3mover.ListObjs(defaultSrcLoca, filt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(objs); i++ {
		obj := osdss3.Object{Size: objs[i].Size, ObjectKey: objs[i].Key, Backend: ""}
		//srcObjs = append(srcObjs, &SourceOject{StorType:defaultSrcLoca.StorType, Obj:&obj})
		srcObjs = append(srcObjs, &obj)
	}
	return srcObjs, nil
}

//to get object details from gcp backend
func getGcpS3Objs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	srcObjs := []*osdss3.Object{}
	objs, err := Gcps3mover.ListObjs(defaultSrcLoca, filt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(objs); i++ {
		obj := osdss3.Object{Size: objs[i].Size, ObjectKey: objs[i].Key, Backend: ""}
		//srcObjs = append(srcObjs, &SourceOject{StorType:defaultSrcLoca.StorType, Obj:&obj})
		srcObjs = append(srcObjs, &obj)
	}
	return srcObjs, nil
}

func getSourceObjs(ctx context.Context, conn *pb.Connector, filt *pb.Filter,
	defaultSrcLoca *LocationInfo) ([]*osdss3.Object, error) {
	switch conn.Type {
	case flowtype.STOR_TYPE_OPENSDS:
		return getOsdsS3Objs(ctx, conn, filt, defaultSrcLoca)
	case flowtype.STOR_TYPE_AWS_S3:
		return getAwsS3Objs(ctx, conn, filt, defaultSrcLoca)
	case flowtype.STOR_TYPE_IBM_COS:
		return getIBMCosObjs(ctx, conn, filt, defaultSrcLoca)
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		return getHwObjs(ctx, conn, filt, defaultSrcLoca)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		return getAzureBlobs(ctx, conn, filt, defaultSrcLoca)
	case flowtype.STOR_TYPE_CEPH_S3:
		return getCephS3Objs(ctx, conn, filt, defaultSrcLoca)
	case flowtype.STOR_TYPE_GCP_S3:
		return getGcpS3Objs(ctx, conn, filt, defaultSrcLoca)
	default:
		{
			logger.Printf("Unsupport storage type:%v\n", conn.Type)
			return nil, errors.New("unsupport storage type")
		}
	}
	return nil, errors.New("Get source objects failed")
}

func prepare4Run(ctx context.Context, j *flowtype.Job, in *pb.RunJobRequest) (srcLoca *LocationInfo, destLoca *LocationInfo,
	objs []*osdss3.Object, err error) {
	srcLoca, err = getConnLocation(ctx, in.SourceConn)
	if err != nil {
		j.Status = flowtype.JOB_STATUS_FAILED
		j.EndTime = time.Now()
		db.DbAdapter.UpdateJob(j)
		logger.Printf("err:%v\n", err)
		return nil, nil, nil, err
	}
	logger.Printf("srcLoca:StorType=%s,VirBucket=%s,BucketName=%s,Region=%s\n",
		srcLoca.StorType, srcLoca.VirBucket, srcLoca.BucketName, srcLoca.Region)
	destLoca, err = getConnLocation(ctx, in.DestConn)
	if err != nil {
		j.Status = flowtype.JOB_STATUS_FAILED
		j.EndTime = time.Now()
		db.DbAdapter.UpdateJob(j)
		return nil, nil, nil, err
	}
	logger.Printf("destLoca:srcLoca:StorType=%s,VirBucket=%s,BucketName=%s,Region=%s\n",
		destLoca.StorType, destLoca.VirBucket, destLoca.BucketName, destLoca.Region)
	logger.Println("Get connector information succeed.")

	//Get Objects which need to be migrated. Calculate the total number and capacity of objects
	objs, err = getSourceObjs(ctx, in.SourceConn, in.GetFilt(), srcLoca)
	totalObjs := len(objs)
	if err != nil {
		logger.Printf("List objects failed, err:%v, total objects:%d\n", err, totalObjs)
		//update database
		j.Status = flowtype.JOB_STATUS_FAILED
		j.EndTime = time.Now()
		db.DbAdapter.UpdateJob(j)
		return srcLoca, destLoca, nil, err
	}
	for i := 0; i < totalObjs; i++ {
		j.TotalCount++
		j.TotalCapacity += objs[i].Size
	}
	if totalObjs == 0 {
		logger.Printf("No data need to migrate. totalObjs=%d, TotalCapacity=%d\n", totalObjs, j.TotalCapacity)
		j.Status = flowtype.JOB_STATUS_SUCCEED
		j.EndTime = time.Now()
		j.Progress = 100
	} else {
		logger.Printf("List objects succeed, total count:%d, total capacity:%d\n", j.TotalCount, j.TotalCapacity)
		j.Status = flowtype.JOB_STATUS_RUNNING
	}
	db.DbAdapter.UpdateJob(j)

	return srcLoca, destLoca, objs, nil
}

func runjob(in *pb.RunJobRequest) error {
	logger.Println("Runjob is called in datamover service.")
	logger.Printf("Request: %+v\n", in)

	j := flowtype.Job{Id: bson.ObjectIdHex(in.Id)}
	j.StartTime = time.Now()

	//TODO:Check if source and destination connectors can access.
	ctx := context.Background()
	_, ok := ctx.Deadline()
	if !ok {
		tmoutCfg, err := strconv.ParseInt(os.Getenv("JOB_MAX_RUN_TIME"), 10, 64)
		if err != nil || tmoutCfg < 60 || tmoutCfg > 2592000 {
			tmoutCfg = int64(JOB_RUN_TIME_MAX)
		}
		durStr := fmt.Sprintf("%ds", tmoutCfg)
		logger.Printf("Vaule of JOB_MAX_RUN_TIME is: %d seconds, durStr:%s.\n", tmoutCfg, durStr)
		tmout, err := time.ParseDuration(durStr)
		if err == nil {
			ctx, _ = context.WithTimeout(ctx, tmout)
		} else {
			logger.Println("Set timeout to the default value.")
			ctx, _ = context.WithTimeout(ctx, 86400*time.Second) //1 day as default
		}
	}

	//prepare for running
	srcLoca, destLoca, objs, err := prepare4Run(ctx, &j, in)
	if err != nil {
		return err
	}
	if j.TotalCount == 0 || j.TotalCapacity == 0 {
		return nil
	}

	//Make channel
	capa := make(chan int64)           //used to transfer capacity of objects
	th := make(chan int, simuRoutines) //concurrent go routines is limited to be  simuRoutines

	//Do migration for each object.
	go doMove(ctx, objs, capa, th, srcLoca, destLoca, in.RemainSource)

	var capacity, count, passedCount, totalObjs int64
	//TODO: What if a part of objects succeed, but the others failed.
	count = 0
	capacity = 0
	passedCount = 0
	totalObjs = j.TotalCount
	tmout := false
	for {
		select {
		case c := <-capa:
			{ //if c equals 0, that means the object is migrated failed.
				count++
				if c >= 0 {
					passedCount++
					capacity += c
				}
				//TODO:update job in database, need to consider the update frequency
				var deci int64 = totalObjs / 10
				if totalObjs < 100 || count == totalObjs || count%deci == 0 {
					//update database
					j.PassedCount = (int64(passedCount))
					j.PassedCapacity = capacity
					j.Progress = int64(capacity * 100 / j.TotalCapacity)
					logger.Printf("capacity:%d,TotalCapacity:%d Progress:%d\n", capacity, j.TotalCapacity, j.Progress)
					db.DbAdapter.UpdateJob(&j)
				}
			}
		case <-time.After(time.Duration(JOB_RUN_TIME_MAX) * time.Second):
			{
				tmout = true
				logger.Println("Timout.")
			}
		}
		if count >= totalObjs || tmout {
			logger.Printf("break, capacity=%d, timout=%v, count=%d, passed count=%d\n", capacity, tmout, count, passedCount)
			close(capa)
			close(th)
			break
		}
	}

	var ret error = nil
	j.PassedCount = int64(passedCount)
	if passedCount < totalObjs {
		errmsg := strconv.FormatInt(totalObjs, 10) + " objects, passed " + strconv.FormatInt(passedCount, 10)
		logger.Printf("Run job failed: %s\n", errmsg)
		ret = errors.New("failed")
		j.Status = flowtype.JOB_STATUS_FAILED
	} else {
		j.Status = flowtype.JOB_STATUS_SUCCEED
	}

	j.EndTime = time.Now()
	for i := 1; i <= 3; i++ {
		err := db.DbAdapter.UpdateJob(&j)
		if err == nil {
			break
		}
		if i == 3 {
			logger.Printf("Update the finish status of job in database failed three times, no need to try more.")
		}
	}

	return ret
}
