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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	c "github.com/opensds/multi-cloud/api/pkg/context"
	. "github.com/opensds/multi-cloud/api/pkg/utils/constants"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	"github.com/opensds/multi-cloud/yigs3/pkg/model"
	"github.com/opensds/multi-cloud/yigs3/proto"
	"golang.org/x/net/context"
)

//Convert function from storage tier to storage class for XML format output
func (s *APIService) tier2class(tier int32) (string, error) {
	{
		mutext.Lock()
		defer mutext.Unlock()
		if len(ClassAndTier) == 0 {
			err := s.loadStorageClassDefinition()
			if err != nil {
				log.Logf("load storage classes failed: %v.\n", err)
				return "", err
			}
		}
	}
	className := ""
	for k, v := range ClassAndTier {
		if v == tier {
			className = k
		}
	}
	if className == "" {
		log.Logf("invalid tier: %d\n", tier)
		return "", fmt.Errorf(InvalidTier)
	}
	return className, nil
}

//Function for GET Bucket Lifecycle API
func (s *APIService) BucketLifecycleGet(request *restful.Request, response *restful.Response) {
	bucketName := request.PathParameter("bucketName")
	log.Logf("received request for getting lifecycle of bucket[name=%s].\n", bucketName)

	ctx := context.Background()
	actx := request.Attribute(c.KContext).(*c.Context)
	bucket, err := s.s3Client.GetBucket(ctx, &s3.CommonRequest{Context: actx.ToJson(), Id: bucketName})
	if err != nil {
		log.Logf("get bucket[name=%s] failed, err=%v.\n", bucketName, err)
		response.WriteError(http.StatusInternalServerError, NoSuchBucket.Error())
		return
	}

	lifecycleConf := []s3.LifecycleRule{}
	err = json.Unmarshal([]byte(bucket.LifeCycle), &lifecycleConf)
	if err != nil {
		log.Logf("unmarshal lifecycle of bucket[%s] failed, lifecycle=%s, err=%v.\n", bucketName, bucket.LifeCycle, err)
		response.WriteError(http.StatusInternalServerError, InternalError.Error())
		return
	}

	// convert back to xml struct
	lifecycleConfXml := model.LifecycleConfiguration{}

	// convert lifecycle rule to xml Rule
	for _, lcRule := range lifecycleConf {
		xmlRule := model.Rule{}

		xmlRule.Status = lcRule.Status
		xmlRule.ID = lcRule.Id
		xmlRule.Filter = converts3FilterToRuleFilter(lcRule.Filter)
		xmlRule.AbortIncompleteMultipartUpload = converts3UploadToRuleUpload(lcRule.AbortIncompleteMultipartUpload)
		xmlRule.Transition = make([]model.Transition, 0)

		//Arranging the transition and expiration actions in XML
		for _, action := range lcRule.Actions {
			log.Logf("action is : %v\n", action)

			if action.Name == ActionNameTransition {
				xmlTransition := model.Transition{}
				xmlTransition.Days = action.Days
				xmlTransition.Backend = action.Backend
				className, err := s.tier2class(action.Tier)
				if err == nil {
					xmlTransition.StorageClass = className
				}
				xmlRule.Transition = append(xmlRule.Transition, xmlTransition)
			}
			if action.Name == ActionNameExpiration {
				xmlExpiration := model.Expiration{}
				xmlExpiration.Days = action.Days
				xmlRule.Expiration = append(xmlRule.Expiration, xmlExpiration)
			}
		}
		// append each xml rule to xml array
		lifecycleConfXml.Rule = append(lifecycleConfXml.Rule, xmlRule)
	}

	// marshall the array back to xml format
	err = response.WriteAsXml(lifecycleConfXml)
	if err != nil {
		log.Logf("write lifecycle of bucket[%s] as xml failed, =%s, err=%v.\n", bucketName, bucket.LifeCycle, err)
		response.WriteError(http.StatusInternalServerError, InternalError.Error())
	}
	log.Log("GET lifecycle successful.")
}

func converts3FilterToRuleFilter(filter *s3.LifecycleFilter) model.Filter {
	retFilter := model.Filter{}
	if filter != nil {
		retFilter.Prefix = filter.Prefix
	}
	return retFilter
}

func converts3UploadToRuleUpload(upload *s3.AbortMultipartUpload) model.AbortIncompleteMultipartUpload {
	retUpload := model.AbortIncompleteMultipartUpload{}
	if upload != nil {
		retUpload.DaysAfterInitiation = upload.DaysAfterInitiation
	}
	return retUpload
}
