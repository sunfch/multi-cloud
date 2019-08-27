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

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	"github.com/micro/go-micro/metadata"
	"github.com/opensds/multi-cloud/api/pkg/common"
	c "github.com/opensds/multi-cloud/api/pkg/context"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	"github.com/opensds/multi-cloud/s3/proto"
	"golang.org/x/net/context"
)

func (s *APIService) ObjectDelete(request *restful.Request, response *restful.Response) {
	url := request.Request.URL
	log.Logf("URL is %v", request.Request.URL.String())

	bucketName := request.PathParameter("bucketName")
	objectKey := request.PathParameter("objectKey")
	if strings.HasSuffix(url.String(), "/") {
		objectKey = objectKey + "/"
	}
	deleteInput := s3.DeleteObjectInput{Key: objectKey, Bucket: bucketName}

	actx := request.Attribute(c.KContext).(*c.Context)
	ctx := metadata.NewContext(context.Background(), map[string]string{
		common.CTX_KEY_USER_ID:   actx.UserId,
		common.CTX_KEY_TENENT_ID: actx.TenantId,
		common.CTX_KEY_IS_ADMIN:  strconv.FormatBool(actx.IsAdmin),
	})

	objectInput := s3.GetObjectInput{Bucket: bucketName, Key: objectKey}
	objectMD, _ := s.s3Client.GetObject(ctx, &objectInput)
	if objectMD != nil {
		client := getBackendByName(ctx, s, objectMD.Backend)
		s3err := client.DELETE(&deleteInput, ctx)
		if s3err.Code != ERR_OK {
			response.WriteError(http.StatusInternalServerError, s3err.Error())
			return
		}
		res, err := s.s3Client.DeleteObject(ctx, &deleteInput)
		if err != nil {
			log.Logf("err is %v\n", err)
			response.WriteError(http.StatusInternalServerError, err)
			return
		}
		log.Logf("Delete object %s successfully.", objectKey)
		response.WriteEntity(res)
	} else {
		log.Logf("No such object")
		return
	}

}
