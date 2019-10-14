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
	"bytes"
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"net/url"
	"time"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
	"github.com/opensds/multi-cloud/api/pkg/common"
	"github.com/opensds/multi-cloud/s3/error"
	pb "github.com/opensds/multi-cloud/s3/proto"
)


//ObjectPut -
func (s *APIService) ObjectCopy(request *restful.Request, response *restful.Response) {
	log.Infoln("enter ObjectCopy")
	copySouce := request.HeaderParameter(common.REQUEST_HEADER_COPY_SOURCE)
	if copySouce != "" {
		response.WriteError(http.StatusBadRequest, s3error.ErrInvalidCopySource)
	}

	targetBucketName := request.PathParameter(common.REQUEST_PATH_BUCKET_NAME)
	targetObjectName := request.PathParameter(common.REQUEST_PATH_OBJECT_KEY)
	//backendName := request.HeaderParameter(common.REQUEST_HEADER_STORAGE_CLASS)
	log.Infof("received request: Copy object, objectkey=%s, bucketName=%s\n:",
		targetObjectName, targetBucketName)

	// copy source is of form: /bucket-name/object-name?versionId=xxxxxx
	copySource := request.Request.Header.Get("X-Amz-Copy-Source")
	// Skip the first element if it is '/', split the rest.
	if strings.HasPrefix(copySource, "/") {
		copySource = copySource[1:]
	}
	splits := strings.SplitN(copySource, "/", 2)

	// Save sourceBucket and sourceObject extracted from url Path.
	var err error
	var sourceBucketName, sourceObjectName, sourceVersion string
	if len(splits) == 2 {
		sourceBucketName = splits[0]
		sourceObjectName = splits[1]
	}
	// If source object is empty, reply back error.
	if sourceObjectName == "" {
		response.WriteError(http.StatusBadRequest, s3error.ErrInvalidCopySource)
		return
	}

	splits = strings.SplitN(sourceObjectName, "?", 2)
	if len(splits) == 2 {
		sourceObjectName = splits[0]
		if !strings.HasPrefix(splits[1], "versionId=") {
			response.WriteError(http.StatusBadRequest, s3error.ErrInvalidCopySource)
			return
		}
		sourceVersion = strings.TrimPrefix(splits[1], "versionId=")
	}

	// X-Amz-Copy-Source should be URL-encoded
	sourceBucketName, err = url.QueryUnescape(sourceBucketName)
	if err != nil {
		response.WriteError(http.StatusBadRequest, s3error.ErrInvalidCopySource)
		return
	}
	sourceObjectName, err = url.QueryUnescape(sourceObjectName)
	if err != nil {
		response.WriteError(http.StatusBadRequest, s3error.ErrInvalidCopySource)
		return
	}

	var isOnlyUpdateMetadata = false

	if sourceBucketName == targetBucketName && sourceObjectName == targetObjectName {
		if request.Request.Header.Get("X-Amz-Metadata-Directive") == "COPY" {
			response.WriteError(http.StatusBadRequest, s3error.ErrInvalidCopyDest)
			return
		} else if request.Request.Header.Get("X-Amz-Metadata-Directive") == "REPLACE" {
			isOnlyUpdateMetadata = true
		} else {
			response.WriteError(http.StatusBadRequest, s3error.ErrInvalidRequestBody)
			return
		}
	}

	log.Infoln("sourceBucketName", sourceBucketName, "sourceObjectName", sourceObjectName,
		"sourceVersion", sourceVersion)

	ctx := common.InitCtxWithAuthInfo(request)
	sourceObject, err := s.s3Client.GetObjectInfo(ctx, &pb.GetObjectInput{
		Bucket:sourceBucketName,
		Key: sourceObjectName,
	})
	if err != nil {
		log.Errorln("Unable to fetch object info. err:", err)
		//response.WriteError(http.StatusBadRequest, s3error.ErrInvalidRequestBody)
		//WriteErrorResponseWithResource(w, r, err, copySource)
		return
	}

	// Verify before x-amz-copy-source preconditions before continuing with CopyObject.
	//if err = checkObjectPreconditions(response.ResponseWriter, request.Request, sourceObject); err != nil {
	//	//WriteErrorResponse(w, r, err)
	//	//return
	//}

	//TODO: In a versioning-enabled bucket, you cannot change the storage class of a specific version of an object. When you copy it, Amazon S3 gives it a new version ID.
	//storageClassFromHeader, err := getStorageClassFromHeader(r)
	//if err != nil {
	//	WriteErrorResponse(w, r, err)
	//	return
	//}
	//if storageClassFromHeader == meta.ObjectStorageClassGlacier || storageClassFromHeader == meta.ObjectStorageClassDeepArchive {
	//	WriteErrorResponse(w, r, ErrInvalidCopySourceStorageClass)
	//	return
	//}

	////if source == dest and X-Amz-Metadata-Directive == REPLACE, only update the meta;
	if isOnlyUpdateMetadata {

	}

	// Note that sourceObject and targetObject are pointers
	targetObject := &pb.Object{}
	targetObject.Acl = sourceObject.Acl
	targetObject.BucketName = targetBucketName
	targetObject.ObjectKey = targetObjectName
	targetObject.Size = sourceObject.Size
	targetObject.Etag = sourceObject.Etag
	targetObject.ContentType = sourceObject.ContentType
	targetObject.CustomAttributes = sourceObject.CustomAttributes

	log.Infoln("srcbucket:", sourceBucketName, " srcObject:", sourceObjectName, " targetBucketName:", targetBucketName, " targetObjectName:", targetObjectName)
	rsp, err := s.s3Client.CopyObject(ctx, &pb.CopyObjectRequest{
		SrcBucket: sourceBucketName,
		TargetBucket: targetBucketName,
		SrcObject: sourceObjectName,
		TargetObject:targetObjectName,
	})
	if err != nil {
		log.Errorln("Unable to copy object from ",
			sourceObjectName, " to ", targetObjectName, " err:", err)
		//WriteErrorResponse(w, r, err)
		return
	}

	//response := GenerateCopyObjectResponse(result.Md5, result.LastModified)
	//encodedSuccessResponse := EncodeResponse(response)
	// write headers
	//if result.Md5 != "" {
	//	w.Header()["ETag"] = []string{"\"" + result.Md5 + "\""}
	//}
	//if sourceVersion != "" {
	//	w.Header().Set("x-amz-copy-source-version-id", sourceVersion)
	//}
	//if result.VersionId != "" {
	//	w.Header().Set("x-amz-version-id", result.VersionId)
	//}
	// Set SSE related headers
	//for _, headerName := range []string{
	//	"X-Amz-Server-Side-Encryption",
	//	"X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id",
	//	"X-Amz-Server-Side-Encryption-Customer-Algorithm",
	//	"X-Amz-Server-Side-Encryption-Customer-Key-Md5",
	//} {
	//	if header := r.Header.Get(headerName); header != "" {
	//		w.Header().Set(headerName, header)
	//	}
	//}
	// write success response.
	//WriteSuccessResponse(w, encodedSuccessResponse)
	// Explicitly close the reader, to avoid fd leaks.
	//pipeReader.Close()

	response.WriteEntity(rsp)

	log.Info("COPY object successfully.")
}

// Encodes the response headers into XML format.
func EncodeResponse(response interface{}) []byte {
	var bytesBuffer bytes.Buffer
	bytesBuffer.WriteString(xml.Header)
	e := xml.NewEncoder(&bytesBuffer)
	e.Encode(response)
	return bytesBuffer.Bytes()
}

// CopyObjectResponse container returns ETag and LastModified of the
// successfully copied object
type CopyObjectResponse struct {
	XMLName      xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ CopyObjectResult" json:"-"`
	ETag         string
	LastModified string // time string of format "2006-01-02T15:04:05.000Z"
}

// GenerateCopyObjectResponse
func GenerateCopyObjectResponse(etag string, lastModified time.Time) CopyObjectResponse {
	return CopyObjectResponse{
		ETag:         "\"" + etag + "\"",
		LastModified: lastModified.UTC().Format(timeFormatAMZ),
	}
}

// WriteSuccessResponse write success headers and response if any.
func WriteSuccessResponse(w http.ResponseWriter, response []byte) {
	if response == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	//ResponseRecorder
	//w.(*ResponseRecorder).status = http.StatusOK
	//w.(*ResponseRecorder).size = int64(len(response))

	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(http.StatusOK)
	w.Write(response)
	w.(http.Flusher).Flush()
}

