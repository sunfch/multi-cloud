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
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	"github.com/micro/go-micro/metadata"
	"github.com/opensds/multi-cloud/api/pkg/common"
	c "github.com/opensds/multi-cloud/api/pkg/context"
	"github.com/opensds/multi-cloud/api/pkg/policy"
	"github.com/opensds/multi-cloud/s3/pkg/model"
	s3 "github.com/opensds/multi-cloud/s3/proto"
	"golang.org/x/net/context"
)

func (s *APIService) GetStorageClasses(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "storageclass:get") {
		return
	}
	log.Log("Received request for storage classes.")

	actx := request.Attribute(c.KContext).(*c.Context)
	ctx := metadata.NewContext(context.Background(), map[string]string{
		common.CTX_KEY_USER_ID:   actx.UserId,
		common.CTX_KEY_TENENT_ID: actx.TenantId,
		common.CTX_KEY_IS_ADMIN:  strconv.FormatBool(actx.IsAdmin),
	})

	res, err := s.s3Client.GetStorageClasses(ctx, &s3.BaseRequest{})
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	tmp := model.ListStorageClasses{}
	classes := []model.StorageClass{}
	for _, v := range res.Classes {
		classes = append(classes, model.StorageClass{Name: v.Name, Tier: v.Tier})
	}
	tmp.Classes = classes

	xmlstring, err := xml.MarshalIndent(tmp, "", "  ")
	if err != nil {
		log.Logf("parse ListStorageClasses error: %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
	} else {
		xmlstring = []byte(xml.Header + string(xmlstring))
		response.Write(xmlstring)
		log.Logf("Get List of storage classes successfully:%v\n", string(xmlstring))
	}
}
