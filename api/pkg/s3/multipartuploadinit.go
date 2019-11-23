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

package s3

import (
	"github.com/emicklei/go-restful"
	"github.com/opensds/multi-cloud/api/pkg/common"
	pb "github.com/opensds/multi-cloud/s3/proto"
	log "github.com/sirupsen/logrus"

	. "github.com/opensds/multi-cloud/s3/error"
)

//ObjectPut -
func (s *APIService) MultiPartUploadInit(request *restful.Request, response *restful.Response) {
	bucketName := request.PathParameter("bucketName")
	objectKey := request.PathParameter("objectKey")

	log.Infof("received request: multipart init, objectkey=%s, bucketName=%s\n:",
		objectKey, bucketName)

	if !isValidObjectName(objectKey) {
		WriteErrorResponse(response, request, ErrInvalidObjectName)
		return
	}

	acl, err := getAclFromHeader(request)
	if err != nil {
		WriteErrorResponse(response, request, err)
		return
	}

	// Save metadata.
	attr := extractMetadataFromHeader(request)

	storageClass, err := getStorageClassFromHeader(request)
	if err != nil {
		WriteErrorResponse(response, request, err)
		return
	}

	ctx := common.InitCtxWithAuthInfo(request)
	result, err := s.s3Client.InitMultipartUpload(ctx, &pb.InitMultiPartRequest{
		BucketName: bucketName, ObjectKey: objectKey, Acl: &pb.Acl{CannedAcl: acl.CannedAcl}, StorageClass: uint32(storageClass), Attrs: attr})
	if err != nil || result.ErrorCode != int32(ErrNoErr) {
		log.Errorln("unable to init multipart. err:", err)
		WriteErrorResponse(response, request, GetFinalError(err, result.ErrorCode))
		return
	}

	data := GenerateInitiateMultipartUploadResponse(bucketName, objectKey, result.UploadID)
	encodedSuccessResponse := EncodeResponse(data)
	// write success response.
	WriteSuccessResponse(response, encodedSuccessResponse)

	log.Infof("Init multipart upload[bucketName=%s, objectKey=%s] successfully.\n",
		bucketName, objectKey)
}
