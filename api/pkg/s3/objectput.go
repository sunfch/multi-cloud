// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
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
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/api/pkg/s3/datastore"

	//	"github.com/micro/go-micro/errors"
	"strconv"

	"github.com/opensds/multi-cloud/api/pkg/policy"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	"github.com/opensds/multi-cloud/s3/proto"
	"golang.org/x/net/context"
)

//ObjectPut -
func (s *APIService) ObjectPut(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "object:put") {
		return
	}
	url := request.Request.URL
	log.Logf("URL is %v", request.Request.URL.String())
	log.Logf("request  is %v\n", request)
	bucketName := request.PathParameter("bucketName")
	log.Logf("bucketName is %v\n:", bucketName)
	objectKey := request.PathParameter("objectKey")
	if strings.HasSuffix(url.String(), "/") {
		objectKey = objectKey + "/"
	}
	log.Logf("objectKey is %v:\n", objectKey)
	contentLenght := request.HeaderParameter("content-length")
	backendName := request.HeaderParameter("x-amz-storage-class")
	/*sc := request.HeaderParameter("x-amz-storage-class")
	if sc == "" { // If storage class is not specified, then STANDARD will be used.
		sc = "STANDARD"
	}
	backendName := request.HeaderParameter("x-backend-name")*/
	log.Logf("backendName is :%v\n", backendName)
	object := s3.Object{}
	object.BucketName = bucketName
	size, _ := strconv.ParseInt(contentLenght, 10, 64)
	log.Logf("object.size is %v\n", size)
	object.Size = size
	object.IsDeleteMarker = ""
	object.InitFlag = ""
	object.LastModified = time.Now().Unix()
	//object.StorageClass = sc
	ctx := context.WithValue(request.Request.Context(), "operation", "upload")

	log.Logf("Received request for create bucket: %s", bucketName)

	log.Logf("objectKey is %v:\n", objectKey)
	object.ObjectKey = objectKey
	var client datastore.DataStoreAdapter
	if backendName != "" {
		object.Backend = backendName
		client = getBackendByName(s, backendName)
	} else {
		bucket, _ := s.s3Client.GetBucket(ctx, &s3.Bucket{Name: bucketName})
		object.Backend = bucket.Backend
		client = getBackendClient(s, bucketName)
	}

	if client == nil {
		response.WriteError(http.StatusInternalServerError, NoSuchBackend.Error())
		return
	}
	log.Logf("enter the PUT method")
	s3err := client.PUT(request.Request.Body, &object, ctx)
	log.Logf("LastModified is %v\n", object.LastModified)

	if s3err != NoError {
		response.WriteError(http.StatusInternalServerError, s3err.Error())
		return
	}
	res, err := s.s3Client.CreateObject(ctx, &object)
	if err != nil {
		log.Logf("err is %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	log.Logf("object.size2  = %v \n", object.Size)
	log.Log("Upload object successfully.")
	response.WriteEntity(res)
}
