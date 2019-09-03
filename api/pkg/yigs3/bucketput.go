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
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	c "github.com/opensds/multi-cloud/api/pkg/context"
	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	. "github.com/opensds/multi-cloud/yigs3/pkg/exception"
	"github.com/opensds/multi-cloud/yigs3/pkg/iam/common"
	"github.com/opensds/multi-cloud/yigs3/pkg/model"
	"github.com/opensds/multi-cloud/yigs3/proto"
	"golang.org/x/net/context"
	//"github.com/opensds/multi-cloud/yigs3/pkg/error"
)

func (s *APIService) BucketPut(request *restful.Request, response *restful.Response) {
	bucketName := strings.ToLower(request.PathParameter("bucketName"))
	if !isValidBucketName(bucketName) {
		//WriteErrorResponse(response, request, error.ErrInvalidBucketName)
		return
	}
	log.Logf("received request: PUT bucket[name=%s]\n", bucketName)

	if len(request.Request.Header.Get("Content-Length")) == 0 {
		log.Logf("Content Length is null!")
		//WriteErrorResponse(response, request, error.ErrInvalidHeader)
		return
	}

	actx := request.Attribute(c.KContext).(*c.Context)
	bucket := s3.Bucket{Name: bucketName}
	bucket.OwnerId = actx.TenantId
	bucket.Deleted = false
	bucket.CreateTime = time.Now().Unix()

	body := ReadBody(request)
	if body != nil {
		createBucketConf := model.CreateBucketConfiguration{}
		err := xml.Unmarshal(body, &createBucketConf)
		if err != nil {
			response.WriteError(http.StatusInternalServerError, err)
			return
		} else {
			backendName := createBucketConf.LocationConstraint
			if backendName != "" {
				log.Logf("backendName is %v\n", backendName)
				bucket.DefaultLocation = backendName
				actx := request.Attribute(c.KContext).(*c.Context).ToJson()
				flag := s.isBackendExist(context.Background(), actx, backendName)
				if flag == false {
					response.WriteError(http.StatusBadRequest, NoSuchBackend.Error())
					return
				}
			} else {
				log.Log("default backend is not provided.")
				response.WriteError(http.StatusBadRequest, NoSuchBackend.Error())
				return
			}
		}
	}

	ctx := context.Background()
	credential := common.Credential{UserId:bucket.OwnerId}
	ctx = context.WithValue(ctx, RequestContextKey, RequestContext{Credential: &credential})
	res, err := s.s3Client.CreateBucket(ctx, &bucket)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	log.Log("Create bucket successfully.")
	response.WriteEntity(res)
}
