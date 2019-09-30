package service

import (
	"context"

	"github.com/micro/go-log"
	pb "github.com/opensds/multi-cloud/s3/proto"
)

func (s *s3Service) ListObjects(ctx context.Context, in *pb.ListObjectsRequest, out *pb.ListObjectResponse) error {
	log.Log("ListObject is called in s3 service.")

	return nil
}

func (s *s3Service) CreateObject(ctx context.Context, in *pb.Object, out *pb.BaseResponse) error {
	log.Log("CreateObject is called in s3 service.")

	return nil
}

func (s *s3Service) UpdateObject(ctx context.Context, in *pb.Object, out *pb.BaseResponse) error {
	log.Log("PutObject is called in s3 service.")

	return nil
}

func (s *s3Service) PutObject(ctx context.Context, in pb.S3_PutObjectStream) error {
	log.Log("PutObject is called in s3 service.")

	return nil
}

func (s *s3Service) GetObject(ctx context.Context, in *pb.GetObjectInput, out *pb.Object) error {
	log.Log("GetObject is called in s3 service.")

	return nil
}

func (s *s3Service) DeleteObject(ctx context.Context, in *pb.DeleteObjectInput, out *pb.BaseResponse) error {
	log.Log("DeleteObject is called in s3 service.")

	return nil
}
