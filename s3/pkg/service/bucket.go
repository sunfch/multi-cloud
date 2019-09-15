package service

import (
	"context"

	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/api/pkg/s3"
	. "github.com/opensds/multi-cloud/s3/error"
	"github.com/opensds/multi-cloud/s3/pkg/helper"
	"github.com/opensds/multi-cloud/s3/pkg/meta/types"
	pb "github.com/opensds/multi-cloud/s3/proto"
)

func (b *s3Service) ListBuckets(ctx context.Context, in *pb.BaseRequest, out *pb.ListBucketsResponse) error {
	log.Log("ListBuckets is called in s3 service.")

	return nil
}

func (b *s3Service) CreateBucket(ctx context.Context, in *pb.Bucket, out *pb.BaseResponse) error {
	log.Log("CreateBucket is called in s3 service.")

	bucketName := in.Name
	if err := s3.CheckValidBucketName(bucketName); err != nil {
		return err
	}

	//credential := ctx.Value(s3.RequestContextKey).(s3.RequestContext).Credential
	processed, err := b.MetaStorage.Db.CheckAndPutBucket(&types.Bucket{Bucket: in})
	if err != nil {
		helper.Logger.Println(5, "Error making checkandput: ", err)
		return err
	}
	if !processed { // bucket already exists, return accurate message
		/*bucket*/ _, err := b.MetaStorage.GetBucket(bucketName, false)
		if err != nil {
			helper.Logger.Println(5, "Error get bucket: ", bucketName, ", with error", err)
			return ErrBucketAlreadyExists
		}
		/*if bucket.OwnerId == credential.UserId {
			return ErrBucketAlreadyOwnedByYou
		} else {
			return ErrBucketAlreadyExists
		}*/
	}
	/*err = b.MetaStorage.Db.AddBucketForUser(bucketName, in.OwnerId)
	if err != nil { // roll back bucket table, i.e. remove inserted bucket
		helper.Logger.Println(5, "Error AddBucketForUser: ", err)
		err = b.MetaStorage.Db.DeleteBucket(&types.Bucket{Bucket: in})
		if err != nil {
			helper.Logger.Println(5, "Error deleting: ", err)
			helper.Logger.Println(5, "Leaving junk bucket unremoved: ", bucketName)
			return err
		}
	}

	if err == nil {
		b.MetaStorage.Cache.Remove(redis.UserTable, meta.BUCKET_CACHE_PREFIX, in.OwnerId)
	}*/

	return err
}

func (b *s3Service) GetBucket(ctx context.Context, in *pb.BaseRequest, out *pb.Bucket) error {
	log.Logf("GetBucket %s is called in s3 service.", in.Id)

	return nil
}

func (b *s3Service) DeleteBucket(ctx context.Context, in *pb.Bucket, out *pb.BaseResponse) error {
	log.Log("DeleteBucket is called in s3 service.")

	return nil
}