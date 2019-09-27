package common

import (
	"context"
	"github.com/micro/go-log"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
)

func OsdsS3CopyObj(ctx context.Context, s3client osdss3.S3Service, obj *osdss3.Object, dstBucket string) error {
	req := osdss3.CopyObjectRequest{
		SrcBucket: obj.BucketName,
		ObjKey: obj.ObjectKey,
		TargetBucket: dstBucket,
	}

	_, err := s3client.CopyObject(ctx, &req)
	if err != nil {
		log.Infof("copy object based on osds s3 failed, obj=%s, err:%v\n", obj.ObjectKey, err)
	}

	return err
}

func OsdsS3DeleteObj(ctx context.Context, s3client osdss3.S3Service, obj *osdss3.Object) error {
	req := osdss3.DeleteObjectInput{
		Key: obj.ObjectKey,
		Bucket: obj.BucketName,
	}

	_, err := s3client.DeleteObject(ctx, &req)
	if err != nil {
		log.Infof("delete object based on osds s3 failed, obj=%s, err:%v\n", obj.ObjectKey, err)
	}

	return nil
}