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

package service

import (
	"context"

	"github.com/opensds/multi-cloud/api/pkg/s3"
	. "github.com/opensds/multi-cloud/s3/error"
	. "github.com/opensds/multi-cloud/s3/pkg/meta/types"
	"github.com/opensds/multi-cloud/s3/pkg/meta/util"
	pb "github.com/opensds/multi-cloud/s3/proto"
	log "github.com/sirupsen/logrus"
)

func (s *s3Service) ListBuckets(ctx context.Context, in *pb.BaseRequest, out *pb.ListBucketsResponse) error {
	log.Info("ListBuckets is called in s3 service.")
	var err error
	defer func() {
		out.ErrorCode = GetErrCode(err)
	}()

	buckets, err := s.MetaStorage.Db.GetBuckets(ctx)
	if err != nil {
		log.Errorf("list buckets failed, err:%v\n", err)
		return nil
	}

	// TODO: paging list
	for j := 0; j < len(buckets); j++ {
		if buckets[j].Deleted != true {
			out.Buckets = append(out.Buckets, &pb.Bucket{
				Name:            buckets[j].Name,
				TenantId:        buckets[j].TenantId,
				CreateTime:      buckets[j].CreateTime,
				Usages:          buckets[j].Usages,
				Tier:            buckets[j].Tier,
				DefaultLocation: buckets[j].DefaultLocation,
			})
		}
	}

	log.Infof("out.Buckets:%+v\n", out.Buckets)
	return nil
}

func (s *s3Service) CreateBucket(ctx context.Context, in *pb.Bucket, out *pb.BaseResponse) error {
	log.Infof("CreateBucket is called in s3 service, in:%+v\n", in)
	var err error
	defer func() {
		out.ErrorCode = GetErrCode(err)
	}()

	bucketName := in.Name
	if err := s3.CheckValidBucketName(bucketName); err != nil {
		err = ErrInvalidBucketName
		return nil
	}

	processed, err := s.MetaStorage.Db.CheckAndPutBucket(ctx, &Bucket{Bucket: in})
	if err != nil {
		log.Error("Error making checkandput: ", err)
		return nil
	}
	log.Infof("create bucket[%s] in database succeed, processed=%v.\n", in.Name, processed)
	if !processed { // bucket already exists, return accurate message
		/*bucket*/ _, err := s.MetaStorage.GetBucket(ctx, bucketName, false)
		if err == nil {
			log.Error("Error get bucket: ", bucketName, ", with error", err)
			err = ErrBucketAlreadyExists
		}
	}

	return nil
}

func (s *s3Service) GetBucket(ctx context.Context, in *pb.Bucket, out *pb.GetBucketResponse) error {
	log.Infof("GetBucket %s is called in s3 service.", in.Id)
	var err error
	defer func() {
		out.ErrorCode = GetErrCode(err)
	}()

	bucket, err := s.MetaStorage.GetBucket(ctx, in.Name, false)
	if err != nil {
		log.Errorf("get bucket[%s] failed, err:%v\n", in.Name, err)
		return nil
	}

	out.BucketMeta = &pb.Bucket{
		Id:              bucket.Id,
		Name:            bucket.Name,
		TenantId:        bucket.TenantId,
		UserId:          bucket.UserId,
		Acl:             bucket.Acl,
		CreateTime:      bucket.CreateTime,
		Deleted:         bucket.Deleted,
		DefaultLocation: bucket.DefaultLocation,
		Tier:            bucket.Tier,
		Usages:          bucket.Usages,
	}

	return nil
}

func (s *s3Service) DeleteBucket(ctx context.Context, in *pb.Bucket, out *pb.BaseResponse) error {
	bucketName := in.Name
	log.Infof("DeleteBucket is called in s3 service, bucketName is %s.\n", bucketName)
	var err error
	defer func() {
		out.ErrorCode = GetErrCode(err)
	}()

	bucket, err := s.MetaStorage.GetBucket(ctx, bucketName, false)
	if err != nil {
		log.Errorf("get bucket failed, err:%+v\n", err)
		return nil
	}

	// Check if bucket is empty
	objs, _, err := s.MetaStorage.Db.ListObjects(ctx, bucketName, false, 1, nil)
	if err != nil {
		log.Errorf("list objects failed, err:%v\n", err)
		return nil
	}
	if len(objs) != 0 {
		log.Errorf("bucket[%s] is not empty.\n", bucketName)
		err = ErrBucketNotEmpty
		return nil
	}
	err = s.MetaStorage.Db.DeleteBucket(ctx, bucket)
	if err != nil {
		log.Errorf("delete bucket[%s] failed, err:%v\n", bucketName, err)
		return nil
	}

	log.Infof("delete bucket[%s] successfully\n", bucketName)
	return nil
}

func (s *s3Service) PutBucketLifecycle(ctx context.Context, in *pb.PutBucketLifecycleRequest, out *pb.BaseResponse) error {
	log.Infof("set lifecycle for bucket[%s]\n", in.BucketName)
	var err error
	defer func() {
		out.ErrorCode = GetErrCode(err)
	}()

	_, tenantId, err := util.GetCredentialFromCtx(ctx)
	if err != nil {
		log.Error("get tenant id failed.")
		return nil
	}

	bucket, err := s.MetaStorage.GetBucket(ctx, in.BucketName, true)
	if err != nil {
		log.Errorf("get bucket failed, err:%v\n", err)
		return nil
	}
	if bucket.TenantId != tenantId {
		log.Errorf("access forbidden, bucket.TenantId=%s, tenantId=%s\n", bucket.TenantId, tenantId)
		err = ErrBucketAccessForbidden
		return nil
	}
	bucket.LifecycleConfiguration = in.Lc
	err = s.MetaStorage.Db.PutBucket(ctx, bucket)
	/* TODO: enable cache, see https://github.com/opensds/multi-cloud/issues/698
	if err == nil {
		s.MetaStorage.Cache.Remove(redis.BucketTable, meta.BUCKET_CACHE_PREFIX, bucketName)
	}*/

	return nil
}

func (s *s3Service) GetBucketLifecycle(ctx context.Context, in *pb.BaseRequest, out *pb.GetBucketLifecycleResponse) error {
	var err error
	defer func() {
		out.ErrorCode = GetErrCode(err)
	}()

	_, tenantId, err := util.GetCredentialFromCtx(ctx)
	if err != nil {
		log.Error("get tenant id failed.")
		return nil
	}

	bucket, err := s.MetaStorage.GetBucket(ctx, in.Id, true)
	if err != nil {
		log.Errorf("get bucket failed, err:%v\n", err)
		return nil
	}

	if bucket.TenantId != tenantId {
		log.Errorf("access forbidden, bucket.TenantId=%s, tenantId=%s\n", bucket.TenantId, tenantId)
		err = ErrBucketAccessForbidden
		return nil
	}

	if len(bucket.LifecycleConfiguration) == 0 {
		log.Errorf("bucket[%s] has no lifecycle configuration\n", in.Id)
		err = ErrNoSuchBucketLc
	} else {
		out.Lc = bucket.LifecycleConfiguration
	}

	return nil
}

func (s *s3Service) DeleteBucketLifecycle(ctx context.Context, in *pb.BaseRequest, out *pb.BaseResponse) error {
	log.Infof("delete lifecycle for bucket:%s\n", in.Id)
	var err error
	defer func() {
		out.ErrorCode = GetErrCode(err)
	}()

	_, tenantId, err := util.GetCredentialFromCtx(ctx)
	if err != nil {
		log.Error("get tenant id failed.")
		return nil
	}

	bucket, err := s.MetaStorage.GetBucket(ctx, in.Id, true)
	if err != nil {
		log.Errorf("get bucket err: %v\n", err)
		return nil
	}

	if bucket.TenantId != tenantId {
		log.Errorf("access forbidden, bucket.TenantId=%s, tenantId=%s\n", bucket.TenantId, tenantId)
		err = ErrBucketAccessForbidden
		return nil
	}
	bucket.LifecycleConfiguration = nil
	err = s.MetaStorage.Db.PutBucket(ctx, bucket)
	if err != nil {
		log.Errorf("update bucket failed, err: %v\n", err)
		return nil
	} else {
		log.Infof("delete lifecycle for bucket:%s successfully\n", in.Id)
	}

	/* TODO: enable cache
	if err == nil {
		yig.MetaStorage.Cache.Remove(redis.BucketTable, meta.BUCKET_CACHE_PREFIX, bucketName)
	}*/

	return nil
}

// ListBucketLifecycle is used by lifecycle management service, not need to return error code
func (s *s3Service) ListBucketLifecycle(ctx context.Context, in *pb.BaseRequest, out *pb.ListBucketsResponse) error {
	log.Info("ListBucketLifecycle is called in s3 service.")
	//buckets := []pb.Bucket{}
	buckets, err := s.MetaStorage.Db.ListBucketLifecycle(ctx)
	if err != nil {
		log.Errorf("list buckets with lifecycle failed, err:%v\n", err)
		return err
	}

	// TODO: paging list
	for _, v := range buckets {
		if v.Deleted != true {
			out.Buckets = append(out.Buckets, &pb.Bucket{
				Name:                   v.Name,
				DefaultLocation:        v.DefaultLocation,
				LifecycleConfiguration: v.LifecycleConfiguration,
			})
		}
	}

	log.Info("list lifecycle successfully")
	return nil
}

func (s *s3Service) ListBucketUploadRecords(ctx context.Context, in *pb.ListBucketUploadRequest, out *pb.ListBucketUploadResponse) error {
	log.Info("ListBucketUploadRecords is called in s3 service.")
	bucketName := in.BucketName

	_, err := s.MetaStorage.GetBucket(ctx, bucketName, true)
	if err != nil {
		log.Errorln("failed to get bucket meta. err:", err)
		return err
	}

	//isAdmin, tenantId, err := util.GetCredentialFromCtx(ctx)
	//if err != nil && isAdmin == false {
	//	log.Error("get tenant id failed")
	//	err = ErrInternalError
	//	return err
	//}
	//
	//switch bucket.Acl.CannedAcl {
	//case "public-read", "public-read-write":
	//	break
	//case "authenticated-read":
	//	if tenantId == "" {
	//		err = ErrBucketAccessForbidden
	//		return err
	//	}
	//default:
	//	if bucket.TenantId != tenantId {
	//		err = ErrBucketAccessForbidden
	//		return err
	//	}
	//}
	log.Errorln("start call ListMultipartUploads at meta.")
	result, err := s.MetaStorage.Db.ListMultipartUploads(in)
	if err != nil {
		log.Errorln("failed to list multipart uploads in meta storage. err:", err)
		return err
	}
	out.Result = result

	log.Infoln("upload number:", len(out.Result.Uploads))

	log.Infoln("List bucket multipart uploads successfully.")
	return nil
}
