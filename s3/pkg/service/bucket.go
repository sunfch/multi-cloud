package service

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/opensds/multi-cloud/api/pkg/s3"
	"github.com/opensds/multi-cloud/s3/pkg/meta/types"
	pb "github.com/opensds/multi-cloud/s3/proto"
	. "github.com/opensds/multi-cloud/s3/error"
)

func (b *s3Service) ListBuckets(ctx context.Context, in *pb.BaseRequest, out *pb.ListBucketsResponse) error {
	log.Info("ListBuckets is called in s3 service.")
	//buckets := []pb.Bucket{}
	buckets, err := b.MetaStorage.Db.GetBuckets(ctx)
	if err != nil {
		log.Errorf("list buckets failed, err:%v\n", err)
		return err
	}

	// TODO: paging list
	for j := 0; j < len(buckets); j++ {
		if buckets[j].Deleted != true {
			out.Buckets = append(out.Buckets, &pb.Bucket{
				Name: buckets[j].Name,
				TenantId: buckets[j].TenantId,
				CreateTime: buckets[j].CreateTime,
				Usages: buckets[j].Usages,
				Tier: buckets[j].Tier,
				DefaultLocation: buckets[j].DefaultLocation,
			})
		}
	}

	return nil
}

func (b *s3Service) CreateBucket(ctx context.Context, in *pb.Bucket, out *pb.BaseResponse) error {
	log.Info("CreateBucket is called in s3 service.")

	bucketName := in.Name
	if err := s3.CheckValidBucketName(bucketName); err != nil {
		return err
	}

	//credential := ctx.Value(s3.RequestContextKey).(s3.RequestContext).Credential
	processed, err := b.MetaStorage.Db.CheckAndPutBucket(ctx, &types.Bucket{Bucket: in})
	if err != nil {
		log.Error("Error making checkandput: ", err)
		return err
	}
	log.Infof("create bucket[%s] in database successfully.\n", in.Name)
	if !processed { // bucket already exists, return accurate message
		/*bucket*/ _, err := b.MetaStorage.GetBucket(ctx, bucketName, false)
		if err != nil {
			log.Error("Error get bucket: ", bucketName, ", with error", err)
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
		log.Error("Error AddBucketForUser: ", err)
		err = b.MetaStorage.Db.DeleteBucket(&types.Bucket{Bucket: in})
		if err != nil {
			log.Error("Error deleting: ", err)
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
	log.Infof("GetBucket %s is called in s3 service.", in.Id)

	bucket, err := b.MetaStorage.GetBucket(ctx, in.Id, false)
	if err != nil {
		log.Error("Error get bucket: ", in.Id, ", with error", err)
		return err
	}

	out = &pb.Bucket{
		Id: bucket.Id,
		Name: bucket.Name,
		TenantId: bucket.TenantId,
		UserId: bucket.UserId,
		Acl: bucket.Acl,
		CreateTime: bucket.CreateTime,
		Deleted: bucket.Deleted,
		DefaultLocation: bucket.DefaultLocation,
		Tier: bucket.Tier,
		Usages: bucket.Usages,
	}

	return nil
}

func (b *s3Service) DeleteBucket(ctx context.Context, in *pb.Bucket, out *pb.BaseResponse) error {
	log.Info("DeleteBucket is called in s3 service.")

	return nil
}
