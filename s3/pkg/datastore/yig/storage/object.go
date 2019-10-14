package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"time"

	. "github.com/opensds/multi-cloud/s3/error"
	dscommon "github.com/opensds/multi-cloud/s3/pkg/datastore/common"
	pb "github.com/opensds/multi-cloud/s3/proto"
	log "github.com/sirupsen/logrus"
)

var latestQueryTime [2]time.Time // 0 is for SMALL_FILE_POOLNAME, 1 is for BIG_FILE_POOLNAME
const CLUSTER_MAX_USED_SPACE_PERCENT = 85

func (yig *YigStorage) PickOneClusterAndPool(bucket string, object string, size int64, isAppend bool) (cluster *CephStorage,
	poolName string) {

	var idx int
	if isAppend {
		poolName = BIG_FILE_POOLNAME
		idx = 1
	} else if size < 0 { // request.ContentLength is -1 if length is unknown
		poolName = BIG_FILE_POOLNAME
		idx = 1
	} else if size < BIG_FILE_THRESHOLD {
		poolName = SMALL_FILE_POOLNAME
		idx = 0
	} else {
		poolName = BIG_FILE_POOLNAME
		idx = 1
	}
	var needCheck bool
	queryTime := latestQueryTime[idx]
	if time.Since(queryTime).Hours() > 24 { // check used space every 24 hours
		latestQueryTime[idx] = time.Now()
		needCheck = true
	}
	var totalWeight int
	clusterWeights := make(map[string]int, len(yig.DataStorage))
	for fsid, _ := range yig.DataStorage {
		cluster, err := yig.MetaStorage.GetCluster(fsid, poolName)
		if err != nil {
			log.Debug("Error getting cluster: ", err)
			continue
		}
		if cluster.Weight == 0 {
			continue
		}
		if needCheck {
			pct, err := yig.DataStorage[fsid].GetUsedSpacePercent()
			if err != nil {
				log.Error("Error getting used space: ", err, "fsid: ", fsid)
				continue
			}
			if pct > CLUSTER_MAX_USED_SPACE_PERCENT {
				log.Error("Cluster used space exceed ", CLUSTER_MAX_USED_SPACE_PERCENT, fsid)
				continue
			}
		}
		totalWeight += cluster.Weight
		clusterWeights[fsid] = cluster.Weight
	}
	if len(clusterWeights) == 0 || totalWeight == 0 {
		log.Warn("Error picking cluster from table cluster in DB! Use first cluster in config to write.")
		for _, c := range yig.DataStorage {
			cluster = c
			break
		}
		return
	}
	N := rand.Intn(totalWeight)
	n := 0
	for fsid, weight := range clusterWeights {
		n += weight
		if n > N {
			cluster = yig.DataStorage[fsid]
			break
		}
	}
	return
}

func (yig *YigStorage) GetClusterByFsName(fsName string) (cluster *CephStorage, err error) {
	if c, ok := yig.DataStorage[fsName]; ok {
		cluster = c
	} else {
		err = errors.New("Cannot find specified ceph cluster: " + fsName)
	}
	return
}

// Write path:
//                                           +-----------+
// PUT object/part                           |           |   Ceph
//         +---------+------------+----------+ Encryptor +----->
//                   |            |          |           |
//                   |            |          +-----------+
//                   v            v
//                  SHA256      MD5(ETag)
//
// SHA256 is calculated only for v4 signed authentication
// Encryptor is enabled when user set SSE headers

/* ctx should contain below elements:
 * size: object size.
 * encryptionKey:
 * md5: the md5 put by user for the uploading object.
 */
func (yig *YigStorage) Put(ctx context.Context, stream io.Reader, obj *pb.Object) (result dscommon.PutResult,
	err error) {
	// get size from context.
	val := ctx.Value(dscommon.CONTEXT_KEY_SIZE)
	if val == nil {
		return result, ErrIncompleteBody
	}
	size := val.(int64)
	// md5 provided by user for uploading object.
	userMd5 := ""
	if val = ctx.Value(dscommon.CONTEXT_KEY_MD5); val != nil {
		userMd5 = val.(string)
	}

	md5Writer := md5.New()

	// Limit the reader to its provided size if specified.
	var limitedDataReader io.Reader
	if size > 0 { // request.ContentLength is -1 if length is unknown
		limitedDataReader = io.LimitReader(stream, size)
	} else {
		limitedDataReader = stream
	}

	cephCluster, poolName := yig.PickOneClusterAndPool(obj.BucketName, obj.ObjectId, size, false)
	if cephCluster == nil {
		log.Errorf("failed to pick cluster and pool for(%s, %s), err: %v", obj.BucketName, obj.ObjectId, err)
		return result, ErrInternalError
	}

	objMeta := ObjectMetaInfo{
		Cluster: cephCluster.Name,
		Pool:    poolName,
	}

	metaBytes, err := json.Marshal(objMeta)
	if err != nil {
		log.Errorf("failed to marshal %v for (%s, %s), err: %v", objMeta, obj.BucketName, obj.ObjectId, err)
		return result, ErrInternalError
	}

	// Mapping a shorter name for the object
	oid := cephCluster.GetUniqUploadName()
	dataReader := io.TeeReader(limitedDataReader, md5Writer)

	bytesWritten, err := cephCluster.Put(poolName, oid, dataReader)
	if err != nil {
		log.Errorf("failed to put(%s, %s), err: %v", poolName, oid, err)
		return
	}
	// Should metadata update failed, add `maybeObjectToRecycle` to `RecycleQueue`,
	// so the object in Ceph could be removed asynchronously
	maybeObjectToRecycle := objectToRecycle{
		location: cephCluster.Name,
		pool:     poolName,
		objectId: oid,
	}
	if bytesWritten < size {
		RecycleQueue <- maybeObjectToRecycle
		log.Errorf("failed to write objects, already written(%d), total size(%d)", bytesWritten, size)
		return result, ErrIncompleteBody
	}

	calculatedMd5 := hex.EncodeToString(md5Writer.Sum(nil))
	log.Info("### calculatedMd5:", calculatedMd5, "userMd5:", userMd5)
	if userMd5 != "" && userMd5 != calculatedMd5 {
		RecycleQueue <- maybeObjectToRecycle
		return result, ErrBadDigest
	}

	if err != nil {
		RecycleQueue <- maybeObjectToRecycle
		return
	}

	// set the bytes written.
	result.Written = bytesWritten
	result.ObjectId = oid
	result.Etag = calculatedMd5
	result.UpdateTime = time.Now().Unix()
	result.Meta = string(metaBytes)

	return result, nil
}

func (yig *YigStorage) Get(ctx context.Context, object *pb.Object, start int64, end int64) (io.ReadCloser, error) {
	// get the cluster name and pool name from meta data of object
	objMeta := ObjectMetaInfo{}

	err := json.Unmarshal([]byte(object.StorageMeta), &objMeta)
	if err != nil {
		log.Errorf("failed to unmarshal storage meta (%s) for (%s, %s), err: %v", object.StorageMeta, object.BucketName, object.ObjectId, err)
		return nil, ErrUnmarshalFailed
	}

	cephCluster, ok := yig.DataStorage[objMeta.Cluster]
	if !ok {
		log.Errorf("cannot find ceph cluster(%s) for obj(%s, %s)", objMeta.Cluster, object.BucketName, object.ObjectId)
		return nil, ErrInvalidObjectName
	}

	len := end - start + 1
	reader, err := cephCluster.getReader(objMeta.Pool, object.ObjectId, start, len)
	if err != nil {
		log.Errorf("failed to get ceph reader for pool(%s), obj(%s,%s) with err: %v", objMeta.Pool, object.BucketName, object.ObjectId, err)
		return nil, err
	}

	return reader, nil
}

func (yig *YigStorage) Delete(ctx context.Context, object *pb.DeleteObjectInput) error {
	return errors.New("not implemented.")
}

/*
* target: should contain BucketName, ObjectId, Size, Etag
*
 */

func (yig *YigStorage) Copy(ctx context.Context, stream io.Reader, target *pb.Object) (result dscommon.PutResult, err error) {
	var limitedDataReader io.Reader
	limitedDataReader = io.LimitReader(stream, target.Size)
	cephCluster, poolName := yig.PickOneClusterAndPool(target.BucketName, target.ObjectId, target.Size, false)
	md5Writer := md5.New()
	oid := cephCluster.GetUniqUploadName()

	objMeta := ObjectMetaInfo{
		Cluster: cephCluster.Name,
		Pool:    poolName,
	}

	metaBytes, err := json.Marshal(objMeta)
	if err != nil {
		log.Errorf("failed to marshal %v for (%s, %s), err: %v", objMeta, target.BucketName, target.ObjectId, err)
		return result, ErrInternalError
	}

	dataReader := io.TeeReader(limitedDataReader, md5Writer)
	var bytesWritten int64
	bytesWritten, err = cephCluster.Put(poolName, oid, dataReader)
	if err != nil {
		log.Errorf("failed to write oid[%s] for obj[%s] in bucket[%s] with err: %v", oid, target.ObjectId, target.BucketName, err)
		return result, err
	}
	// Should metadata update failed, add `maybeObjectToRecycle` to `RecycleQueue`,
	// so the object in Ceph could be removed asynchronously
	maybeObjectToRecycle := objectToRecycle{
		location: cephCluster.Name,
		pool:     poolName,
		objectId: oid,
	}
	if bytesWritten < target.Size {
		RecycleQueue <- maybeObjectToRecycle
		return result, ErrIncompleteBody
	}

	calculatedMd5 := hex.EncodeToString(md5Writer.Sum(nil))
	if calculatedMd5 != target.Etag {
		RecycleQueue <- maybeObjectToRecycle
		return result, ErrBadDigest
	}

	result.Etag = calculatedMd5
	result.Written = bytesWritten
	result.ObjectId = oid
	result.UpdateTime = time.Now().Unix()
	result.Meta = string(metaBytes)
	target.ObjectId = oid

	log.Debugf("succeeded to copy object[%s] in bucket[%s] with oid[%s]", target.ObjectId, target.BucketName, oid)
	return result, nil
}

