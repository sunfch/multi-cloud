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
	"log"
	"os"
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
	"github.com/opensds/multi-cloud/datamover/pkg/gcp/s3"
	"github.com/opensds/multi-cloud/datamover/pkg/huawei/obs"
	"github.com/opensds/multi-cloud/datamover/pkg/ibm/cos"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	pb "github.com/opensds/multi-cloud/datamover/proto"
	s3utils "github.com/opensds/multi-cloud/s3/pkg/utils"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	mvtool "github.com/opensds/multi-cloud/datamover/pkg/drivers/https/common"
	"github.com/micro/go-micro/metadata"
	"fmt"
	"github.com/opensds/multi-cloud/api/pkg/common"
)

var simuRoutines = 10
var PART_SIZE int64 = 16 * 1024 * 1024 //The max object size that can be moved directly, default is 16M.
var JOB_RUN_TIME_MAX = 86400           //seconds, equals 1 day
var s3client osdss3.S3Service
var bkendclient backend.BackendService
var logger = log.New(os.Stdout, "", log.LstdFlags)

type MoveReqStaticInfo struct {
	//actx string
	remainSource bool
	srcLoc *LocationInfo
	dstLoc *LocationInfo
}

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
		logger.Printf("unmarshal failed, err:%v\n", err)
		return err
	}

	//Check the status of job, and run it if needed
	status := db.DbAdapter.GetJobStatus(job.Id)
	if status != flowtype.JOB_STATUS_PENDING {
		logger.Printf("job[id#%s] is not in %s status.\n", job.Id, flowtype.JOB_STATUS_PENDING)
		return nil //No need to consume this message again
	}

	logger.Printf("HandleMsg:job=%+v\n", job)
	// TODO: how many jobs can run simutaneously
	go runjob(&job)
	return nil
}

func doMove(ctx context.Context, info *MoveReqStaticInfo, objs []*osdss3.Object, capa chan int64, th chan int) {
	//Only three routines allowed to be running at the same time
	//th := make(chan int, simuRoutines)
	for i := 0; i < len(objs); i++ {
		if objs[i].Tier == s3utils.Tier999 {
			// archived object cannot be moved currently
			logger.Printf("Object(key:%s) is archived, cannot be migrated.\n", objs[i].ObjectKey)
			continue
		}
		logger.Printf("************Begin to move obj(key:%s)\n", objs[i].ObjectKey)

		go move(ctx, info, objs[i], capa, th)

		//Create one routine
		th <- 1
		logger.Printf("doMigrate: produce 1 routine, len(th):%d.\n", len(th))
	}
}

func CopyObj(obj *osdss3.Object, srcLoca *LocationInfo, destLoca *LocationInfo) error {
	logger.Printf("*****move object[%s] from #%s[%s]# to #%s[%s]#, size is %d.\n", obj.ObjectKey,
		srcLoca.BucketName, srcLoca.StorType, destLoca.BucketName, destLoca.StorType, obj.Size)

	if obj.Size <= 0 {
		return nil
	}
	buf := make([]byte, obj.Size)
	var size int64 = 0
	var err error = nil
	var downloader, uploader MoveWorker

	//download
	switch srcLoca.StorType {
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		downloader = &obsmover.ObsMover{}
		size, err = downloader.DownloadObj(obj.ObjectKey, srcLoca, buf)
	case flowtype.STOR_TYPE_AWS_S3:
		downloader = &s3mover.S3Mover{}
		size, err = downloader.DownloadObj(obj.ObjectKey, srcLoca, buf)
	case flowtype.STOR_TYPE_IBM_COS:
		downloader = &ibmcosmover.IBMCOSMover{}
		size, err = downloader.DownloadObj(obj.ObjectKey, srcLoca, buf)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		downloader = &blobmover.BlobMover{}
		size, err = downloader.DownloadObj(obj.ObjectKey, srcLoca, buf)
	case flowtype.STOR_TYPE_CEPH_S3:
		downloader = &cephs3mover.CephS3Mover{}
		size, err = downloader.DownloadObj(obj.ObjectKey, srcLoca, buf)
	case flowtype.STOR_TYPE_GCP_S3:
		downloader = &Gcps3mover.GcpS3Mover{}
		size, err = downloader.DownloadObj(obj.ObjectKey, srcLoca, buf)
	default:
		{
			logger.Printf("not support source backend type:%v\n", srcLoca.StorType)
			err = errors.New("not support source backend type")
		}
	}

	if err != nil {
		logger.Printf("download object[%s] failed.", obj.ObjectKey)
		return err
	}
	logger.Printf("Download object[%s] succeed, size=%d\n", obj.ObjectKey, size)

	//upload
	switch destLoca.StorType {
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		uploader = &obsmover.ObsMover{}
		err = uploader.UploadObj(obj.ObjectKey, destLoca, buf)
	case flowtype.STOR_TYPE_AWS_S3:
		uploader = &s3mover.S3Mover{}
		err = uploader.UploadObj(obj.ObjectKey, destLoca, buf)
	case flowtype.STOR_TYPE_IBM_COS:
		uploader = &ibmcosmover.IBMCOSMover{}
		err = uploader.UploadObj(obj.ObjectKey, destLoca, buf)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		uploader = &blobmover.BlobMover{}
		err = uploader.UploadObj(obj.ObjectKey, destLoca, buf)
	case flowtype.STOR_TYPE_CEPH_S3:
		uploader = &cephs3mover.CephS3Mover{}
		err = uploader.UploadObj(obj.ObjectKey, destLoca, buf)
	case flowtype.STOR_TYPE_GCP_S3:
		uploader = &Gcps3mover.GcpS3Mover{}
		err = uploader.UploadObj(obj.ObjectKey, destLoca, buf)
	default:
		logger.Printf("not support destination backend type:%v\n", destLoca.StorType)
		return errors.New("not support destination backend type.")
	}
	if err != nil {
		logger.Printf("upload object[bucket:%s,key:%s] failed, err:%v.\n", destLoca.BucketName,
			obj.ObjectKey, err)
	} else {
		logger.Printf("upload object[bucket:%s,key:%s] successfully.\n", destLoca.BucketName, obj.ObjectKey)
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
		logger.Printf("unsupport storType[%s] to init multipart download.\n", srcLoca.StorType)
	}

	return nil, errors.New("unsupport storage type.")
}

func multiPartUploadInit(objKey string, destLoca *LocationInfo) (mover MoveWorker, uploadId string, err error) {
	uploadId = ""
	switch destLoca.StorType {
	case flowtype.STOR_TYPE_AWS_S3:
		mover = &s3mover.S3Mover{}
		uploadId, err = mover.MultiPartUploadInit(objKey, destLoca)
		return
	case flowtype.STOR_TYPE_IBM_COS:
		mover = &ibmcosmover.IBMCOSMover{}
		uploadId, err = mover.MultiPartUploadInit(objKey, destLoca)
		return
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		mover = &obsmover.ObsMover{}
		uploadId, err = mover.MultiPartUploadInit(objKey, destLoca)
		return
	case flowtype.STOR_TYPE_AZURE_BLOB:
		mover = &blobmover.BlobMover{}
		uploadId, err = mover.MultiPartUploadInit(objKey, destLoca)
		return
	case flowtype.STOR_TYPE_CEPH_S3:
		mover = &cephs3mover.CephS3Mover{}
		uploadId, err = mover.MultiPartUploadInit(objKey, destLoca)
		return
	case flowtype.STOR_TYPE_GCP_S3:
		mover = &Gcps3mover.GcpS3Mover{}
		uploadId, err = mover.MultiPartUploadInit(objKey, destLoca)
		return mover, uploadId, err
	default:
		logger.Printf("unsupport storType[%s] to download.\n", destLoca.StorType)
	}

	return nil, uploadId, errors.New("unsupport storage type")
}

func abortMultipartUpload(objKey string, destLoca *LocationInfo, mover MoveWorker) error {
	switch destLoca.StorType {
	case flowtype.STOR_TYPE_AWS_S3, flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE,
		flowtype.STOR_TYPE_HW_FUSIONCLOUD, flowtype.STOR_TYPE_AZURE_BLOB, flowtype.STOR_TYPE_CEPH_S3, flowtype.STOR_TYPE_GCP_S3, flowtype.STOR_TYPE_IBM_COS:
		return mover.AbortMultipartUpload(objKey, destLoca)
	default:
		logger.Printf("unsupport storType[%s] to download.\n", destLoca.StorType)
	}

	return errors.New("unsupport storage type")
}

func completeMultipartUpload(objKey string, destLoca *LocationInfo, mover MoveWorker) error {
	switch destLoca.StorType {
	case flowtype.STOR_TYPE_AWS_S3, flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE,
		flowtype.STOR_TYPE_HW_FUSIONCLOUD, flowtype.STOR_TYPE_AZURE_BLOB, flowtype.STOR_TYPE_CEPH_S3, flowtype.STOR_TYPE_GCP_S3, flowtype.STOR_TYPE_IBM_COS:
		return mover.CompleteMultipartUpload(objKey, destLoca)
	default:
		logger.Printf("unsupport storType[%s] to download.\n", destLoca.StorType)
	}

	return errors.New("unsupport storage type")
}

func MultipartCopyObj(obj *osdss3.Object, srcLoca *LocationInfo, destLoca *LocationInfo) error {
	logger.Printf("*****multi-part move object[%s] from #%s[%s]# to #%s[%s]#, size is %d.\n", obj.ObjectKey,
		srcLoca.BucketName, srcLoca.StorType, destLoca.BucketName, destLoca.StorType, obj.Size)

	partCount := int64(obj.Size / PART_SIZE)
	if obj.Size%PART_SIZE != 0 {
		partCount++
	}

	buf := make([]byte, PART_SIZE)
	var i int64
	var err error
	var uploadMover, downloadMover MoveWorker
	var uploadId string
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
		readSize, err := downloadMover.DownloadRange(obj.ObjectKey, srcLoca, buf, start, end)
		if err != nil {
			return errors.New("download failed")
		}
		//fmt.Printf("Download part %d range[%d:%d] successfully.\n", partNumber, offset, end)
		if int64(readSize) != currPartSize {
			logger.Printf("internal error, currPartSize=%d, readSize=%d\n", currPartSize, readSize)
			return errors.New(DMERR_InternalError)
		}

		//upload
		if partNumber == 1 {
			//init multipart upload
			uploadMover, uploadId, err = multiPartUploadInit(obj.ObjectKey, destLoca)
			if err != nil {
				return err
			}
			logger.Printf("init multipart upload successfully, uploadId:%s\n", uploadId)
		}
		err1 := uploadMover.UploadPart(obj.ObjectKey, destLoca, currPartSize, buf, partNumber, offset)
		if err1 != nil {
			err := abortMultipartUpload(obj.ObjectKey, destLoca, uploadMover)
			if err != nil {
				logger.Printf("Abort s3 multipart upload failed, err:%v\n", err)
			}

			return errors.New("multipart upload failed")
		}
	}

	err = completeMultipartUpload(obj.ObjectKey, destLoca, uploadMover)
	if err != nil {
		logger.Println(err.Error())
		err := abortMultipartUpload(obj.ObjectKey, destLoca, uploadMover)
		if err != nil {
			logger.Printf("abort s3 multipart upload failed, err:%v\n", err)
		}
	}

	return err
}

func deleteRemoteObj(ctx context.Context, obj *osdss3.Object, loca *LocationInfo) error {
	objKey := obj.ObjectKey
	/*if loca.VirBucket != "" {
		objKey = loca.VirBucket + "/" + objKey
	}*/
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
		logger.Printf("delete object[objkey:%s] from backend storage failed.\n", obj.ObjectKey)
		err = errors.New(DMERR_UnSupportBackendType)
	}

	if err != nil {
		return err
	}

	return err
}

/*func osdss3Copy(ctx context.Context, info *MoveReqStaticInfo, obj *osdss3.Object) error {
	copyRequest := osdss3.CopyObjectRequest{
		Context: info.actx,
		SrcBucket: obj.BucketName,
		SrcObject: obj.ObjectKey,
		DstBucket: info.dstLoc.BucketName,
	}

	_, err := s3client.CopyObject(ctx, &copyRequest)
	if err != nil {
		log.Printf("osdss3Copy failed, err:%v\n", err)
	}

	return err
}*/

func defaultCopy(ctx context.Context, info *MoveReqStaticInfo, obj *osdss3.Object) error {
	log.Printf("defaultCopy: objKey=%s\n", obj.ObjectKey)

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
		err = CopyObj(obj, info.srcLoc, info.dstLoc)
	} else {
		err = MultipartCopyObj(obj, info.srcLoc, info.dstLoc)
	}

	return err
}

func move(ctx context.Context, info *MoveReqStaticInfo, obj *osdss3.Object, capa chan int64, th chan int) {
	logger.Printf("Obj[%s] is stored in the backend is [%s], default backend is [%s], target backend is [%s].\n",
		obj.ObjectKey)

	var err error
	if info.srcLoc.StorType == flowtype.STOR_TYPE_OPENSDS && info.dstLoc.StorType == flowtype.STOR_TYPE_OPENSDS {
		err = mvtool.OsdsS3CopyObj(ctx, s3client, obj, info.dstLoc.BucketName)
		logger.Printf("remainSource for object[%s] is:%v.", obj.ObjectKey, info.remainSource)
		if err != nil && !info.remainSource {
			mvtool.OsdsS3DeleteObj(ctx, s3client, obj)
			//TODO: what if delete failed
		}
	} else {
		err = defaultCopy(ctx, info, obj)
		// Delete source data if needed
		logger.Printf("remainSource for object[%s] is:%v.", obj.ObjectKey, info.remainSource)
		if err != nil && !info.remainSource {
			deleteRemoteObj(ctx, obj, info.srcLoc)
			//TODO: what if delete failed
		}
	}

	if err != nil {
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

func updateJob(j *flowtype.Job) {
	for i := 1; i <= 3; i++ {
		err := db.DbAdapter.UpdateJob(j)
		if err == nil {
			break
		}
		if i == 3 {
			logger.Printf("update the finish status of job in database failed three times, no need to try more.")
		}
	}
}

func getConnectorInfo(ctx context.Context, in *pb.RunJobRequest, req *MoveReqStaticInfo) error {
	if in.SourceConn.Type != flowtype.STOR_TYPE_OPENSDS && in.DestConn.Type != flowtype.STOR_TYPE_OPENSDS {
		log.Println("move between non-opensds buckets")

		srcLoc, dstLoc, err := getLocationInfo(ctx, in)
		if err != nil {
			log.Printf("get location failed, err:%v\n", err)
			return err
		}

		req.srcLoc.StorType = srcLoc.StorType
		req.srcLoc.BucketName = srcLoc.BucketName
		req.srcLoc.Access = srcLoc.Access
		req.srcLoc.Security = srcLoc.Security
		req.srcLoc.Region = srcLoc.Region
		req.dstLoc.StorType = dstLoc.StorType
		req.dstLoc.BucketName = dstLoc.BucketName
		req.dstLoc.Access = dstLoc.Access
		req.dstLoc.Security = dstLoc.Security
		req.dstLoc.Region = dstLoc.Region
	} else if in.SourceConn.Type == flowtype.STOR_TYPE_OPENSDS && in.DestConn.Type == flowtype.STOR_TYPE_OPENSDS {
		log.Println("move between opensds buckets")

		req.srcLoc.StorType = flowtype.STOR_TYPE_OPENSDS
		req.srcLoc.BucketName = in.SourceConn.BucketName
		req.dstLoc.StorType = flowtype.STOR_TYPE_OPENSDS
		req.dstLoc.BucketName = in.DestConn.BucketName
	} else {
		log.Println("cannot move between non-opensds bucket and opensds bucket")
		return fmt.Errorf("cannot move between non-opensds bucket and opensds bucket")
	}

	return nil
}

func updateJobFinalStatus(ctx context.Context, j *flowtype.Job, passedNum, totalNum int64) error {
	var err error
	j.PassedCount = int64(passedNum)
	if passedNum < totalNum {
		errmsg := strconv.FormatInt(totalNum, 10) + " objects, passed " + strconv.FormatInt(passedNum, 10)
		logger.Printf("run job failed: %s\n", errmsg)
		err = errors.New("failed")
		j.Status = flowtype.JOB_STATUS_FAILED
	} else {
		j.Status = flowtype.JOB_STATUS_SUCCEED
	}

	j.EndTime = time.Now()
	for i := 1; i <= 3; i++ {
		err := db.DbAdapter.UpdateJob(j)
		if err == nil {
			break
		}
		if i == 3 {
			logger.Printf("update the finish status of job in database failed three times, no need to try more.")
		}
	}

	return err
}

func runjob(in *pb.RunJobRequest) error {
	logger.Println("Runjob is called in datamover service.")
	logger.Printf("Request: %+v\n", in)

	// set context tiemout
	ctx := metadata.NewContext(context.Background(), map[string]string{
		common.CTX_KEY_USER_ID:    in.UserId,
		common.CTX_KEY_TENANT_ID:  in.TenanId,
	})
	dur := getCtxTimeout()
	_, ok := ctx.Deadline()
	if !ok {
		ctx, _ = context.WithTimeout(ctx, dur)
	}

	// init job
	j := flowtype.Job{Id: bson.ObjectIdHex(in.Id), StartTime: time.Now(), Status: flowtype.JOB_STATUS_RUNNING}
	updateJob(&j)

	// get connector information
	mvReqStaticInfo := MoveReqStaticInfo{remainSource: in.RemainSource}
	err := getConnectorInfo(ctx, in, &mvReqStaticInfo)
	if err != nil {
		j.Status = flowtype.JOB_STATUS_FAILED
		j.EndTime = time.Now()
		updateJob(&j)
		return err
	}

	// get total count and total size of objects need to be migrated
	err = updateCountInfo(ctx, in, &j)
	if err != nil {
		return err
	}

	// used to transfer capacity(size) of objects
	capa := make(chan int64)
	// concurrent go routines is limited to be simuRoutines
	th := make(chan int, simuRoutines)

	var offset, limit int32 = 0, 1000
	for {
		objs, err := getObjs(ctx, in, offset, limit)
		if err != nil {
			//update database
			j.Status = flowtype.JOB_STATUS_FAILED
			j.EndTime = time.Now()
			db.DbAdapter.UpdateJob(&j)
			return err
		}

		//Do migration for each object.
		//go doMove(ctx, objs, capa, th, srcLoca, destLoca, in.RemainSource)
		go doMove(ctx, &mvReqStaticInfo, objs, capa, th)

		num := len(objs)
		if num < int(limit) {
			break
		}
		offset = offset + int32(num)
	}

	var capacity, count, passedCount, totalNum int64 = 0, 0, 0, j.TotalCount
	tmout := false
	for {
		select {
		case c := <-capa:
			{ //if c is less than 0, that means the object is migrated failed.
				count++
				if c >= 0 {
					passedCount++
					capacity += c
				}

				var deci int64 = totalNum / 10
				if totalNum < 100 || count == totalNum || count%deci == 0 {
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
		if count >= totalNum || tmout {
			logger.Printf("break, capacity=%d, timout=%v, count=%d, passed count=%d\n", capacity, tmout, count, passedCount)
			close(capa)
			close(th)
			break
		}
	}

	return updateJobFinalStatus(ctx, &j, passedCount, totalNum)
}
