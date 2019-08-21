package common

import (
	"context"
	"github.com/micro/go-log"
	osdss3 "github.com/opensds/multi-cloud/yigs3/proto"
)

func OsdsS3CopyObj(ctx context.Context, s3client osdss3.S3Service, obj *osdss3.Object, actx, dstBucket string) error {
	req := osdss3.CopyObjectRequest{
		Context: actx,
		SrcBucket: obj.BucketName,
		ObjKey: obj.ObjectKey,
		TargetBucket: dstBucket,
	}

	_, err := s3client.CopyObject(ctx, &req)
	if err != nil {
		log.Logf("copy object based on osds s3 failed, obj=%s, err:%v\n", obj.ObjectKey, err)
	}

	return err
}


func OsdsS3DeleteObj(ctx context.Context, s3client osdss3.S3Service, obj *osdss3.Object, actx string) error {
	req := osdss3.DeleteObjectInput{
		Context: actx,
		Key: obj.ObjectKey,
		Bucket: obj.BucketName,
	}

	_, err := s3client.DeleteObject(ctx, &req)
	if err != nil {
		log.Logf("delete object based on osds s3 failed, obj=%s, err:%v\n", obj.ObjectKey, err)
	}
	return nil
}