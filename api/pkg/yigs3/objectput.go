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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/api/pkg/utils/constants"
	. "github.com/opensds/multi-cloud/yigs3/pkg/exception"
	"github.com/opensds/multi-cloud/yigs3/pkg/utils"
	"github.com/opensds/multi-cloud/yigs3/proto"
	"golang.org/x/net/context"
)

//ObjectPut -
func (s *APIService) ObjectPut(request *restful.Request, response *restful.Response) {
	bucketName := request.PathParameter("bucketName")
	objectKey := request.PathParameter("objectKey")

	url := request.Request.URL
	if strings.HasSuffix(url.String(), "/") {
		objectKey = objectKey + "/"
	}
	log.Logf("received request: PUT object, objectkey=%s, bucketName=%s\n:", objectKey, bucketName)

	contentLenght := request.HeaderParameter("content-length")
	backendName := request.HeaderParameter("x-amz-storage-class")
	log.Logf("contentLenght=%s, backendName is :%v\n", contentLenght, backendName)

	ctx := context.WithValue(request.Request.Context(), "operation", "upload")

	// check if bucket exist
	bucketMeta := s.getBucketMeta(ctx, bucketName)
	if bucketMeta == nil {
		response.WriteError(http.StatusBadRequest, NoSuchBucket.Error())
		return
	}

	object := s3.Object{}
	object.ObjectKey = objectKey
	object.BucketName = bucketName
	size, _ := strconv.ParseInt(contentLenght, 10, 64)
	log.Logf("object.size is %v\n", size)
	object.Size = size
	object.DeleteMarker = false
	object.LastModifiedTime = time.Now().Unix()
	// Currently, only support tier1 as default
	object.Tier = int32(utils.Tier1)
	// standard as default
	object.StorageClass = constants.StorageClassOpenSDSStandard

	if backendName != "" {
		// check if backend exist
		if s.isBackendExist(ctx, backendName) == false {
			response.WriteError(http.StatusBadRequest, NoSuchBackend.Error())
		}
		object.Location = backendName
	} else {
		object.Location = bucketMeta.DefaultLocation
	}

	// TODO: check size, cause there is a limitation of transfer size for protobuf ...
	// Limit the reader to its provided size if specified.
	/*var limitedDataReader io.Reader
	if size > 0 { // request.ContentLength is -1 if length is unknown
		limitedDataReader = io.LimitReader(request.Request.Body, size)
	} else {
		limitedDataReader = request.Request.Body
	}*/

	res, err := s.s3Client.PutObject(ctx, &s3.PutObjectRequest{
		Meta: &object,
		// TODO: fill data
		//data:
	})
	if err != nil {
		log.Logf("failed to PUT object, err is %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Log("PUT object successfully.")
	response.WriteEntity(res)
}
