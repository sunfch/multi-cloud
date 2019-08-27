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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	"github.com/micro/go-micro/metadata"
	"github.com/opensds/multi-cloud/api/pkg/common"
	c "github.com/opensds/multi-cloud/api/pkg/context"
	"github.com/opensds/multi-cloud/s3/proto"
)

func checkLastmodifiedFilter(fmap *map[string]string) error {
	for k, v := range *fmap {
		if k != "lt" && k != "lte" && k != "gt" && k != "gte" {
			log.Logf("invalid query parameter:k=%s,v=%s\n", k, v)
			return errors.New("invalid query parameter")
		} else {
			_, err := strconv.Atoi(v)
			if err != nil {
				log.Logf("invalid query parameter:k=%s,v=%s, err=%v\n", k, v, err)
				return errors.New("invalid query parameter")
			}
		}
	}

	return nil
}

func checkObjKeyFilter(val string) (string, error) {
	// val should be like: objeKey=like:parttern
	if strings.HasPrefix(val, "like:") == false {
		log.Logf("invalid object key filter:%s\n", val)
		return "", fmt.Errorf("invalid object key filter:%s", val)
	}

	vals := strings.Split(val, ":")
	if len(vals) <= 1 {
		log.Logf("invalid object key filter:%s\n", val)
		return "", fmt.Errorf("invalid object key filter:%s", val)
	}

	var ret string
	for i := 1; i < len(vals); i++ {
		ret = ret + vals[i]
	}

	return ret, nil
}

func (s *APIService) BucketGet(request *restful.Request, response *restful.Response) {
	limit, offset, err := common.GetPaginationParam(request)
	if err != nil {
		log.Logf("get pagination parameters failed: %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	bucketName := request.PathParameter("bucketName")
	log.Logf("Received request for bucket details: %s\n", bucketName)

	filterOpts := []string{common.KObjKey, common.KLastModified}
	filter, err := common.GetFilter(request, filterOpts)
	if err != nil {
		log.Logf("get filter failed: %v\n", err)
		response.WriteError(http.StatusBadRequest, err)
		return
	} else {
		log.Logf("Get filter for BucketGet, filterOpts=%+v, filter=%+v\n",
			filterOpts, filter)
	}

	if filter[common.KObjKey] != "" {
		//filter[common.KObjKey] should be like: like:parttern
		ret, err := checkObjKeyFilter(filter[common.KObjKey])
		if err != nil {
			log.Logf("invalid objkey:%s\v", filter[common.KObjKey])
			response.WriteError(http.StatusBadRequest,
				fmt.Errorf("invalid objkey, it should be like objkey=like:parttern"))
			return
		}
		filter[common.KObjKey] = ret
	}

	// Check validation of query parameter. Example of lastmodified: {"lt":"100", "gt":"30"}
	if filter[common.KLastModified] != "" {
		var tmFilter map[string]string
		err := json.Unmarshal([]byte(filter[common.KLastModified]), &tmFilter)
		if err != nil {
			log.Logf("invalid lastModified:%s\v", filter[common.KLastModified])
			response.WriteError(http.StatusBadRequest,
				fmt.Errorf("invalid lastmodified, it should be like lastmodified={\"lt\":\"numb\"}"))
			return
		}
		err = checkLastmodifiedFilter(&tmFilter)
		if err != nil {
			log.Logf("invalid lastModified:%s\v", filter[common.KLastModified])
			response.WriteError(http.StatusBadRequest,
				fmt.Errorf("invalid lastmodified, it should be like lastmodified={\"lt\":\"numb\"}"))
			return
		}
	}

	req := s3.ListObjectsRequest{
		Bucket: bucketName,
		Filter: filter,
		Offset: offset,
		Limit:  limit,
	}

	actx := request.Attribute(c.KContext).(*c.Context)
	ctx := metadata.NewContext(context.Background(), map[string]string{
		common.CTX_KEY_USER_ID:   actx.UserId,
		common.CTX_KEY_TENENT_ID: actx.TenantId,
		common.CTX_KEY_IS_ADMIN:  strconv.FormatBool(actx.IsAdmin),
	})

	res, err := s.s3Client.ListObjects(ctx, &req)
	log.Logf("list objects result: %v\n", res)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Log("Get bucket successfully.")
	response.WriteEntity(res)
}
