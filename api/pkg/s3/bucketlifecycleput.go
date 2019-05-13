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
	"crypto/md5"
	"errors"
	"fmt"
	"sync"
	"sort"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	. "github.com/opensds/multi-cloud/api/pkg/utils/constants"
	"github.com/opensds/multi-cloud/s3/pkg/model"
	s3 "github.com/opensds/multi-cloud/s3/proto"
	"golang.org/x/net/context"
)

// Map from storage calss to tier
var ClassAndTier map[string]int32
var mutext sync.Mutex

//type InterActions []*InternalLifecycleAction

type InternalLifecycleAction struct {
	Name    string
	Days    int32
	Tier    int32
	Backend string
}

func (s *APIService) loadStorageClassDefinition() error {
	ctx := context.Background()

	log.Log("Load storage classes.")
	res, err := s.s3Client.GetStorageClasses(ctx, &s3.BaseRequest{})
	if err != nil {
		log.Logf("get storage classes from s3 service failed: %v\n", err)
		return err
	}

	ClassAndTier = make(map[string]int32)
	for _, v := range res.Classes {
		ClassAndTier[v.Name] = v.Tier
	}

	return nil
}

func (s *APIService) class2tier(name string) (int32, error) {
	{
		mutext.Lock()
		defer mutext.Unlock()
		if len(ClassAndTier) == 0 {
			err := s.loadStorageClassDefinition()
			if err != nil {
				log.Logf("load storage classes failed: %v.\n", err)
				return 0, err
			}
		}
	}

	tier, ok := ClassAndTier[name]
	if !ok {
		log.Logf("translate storage class name[%s] to tier failed: %s.\n", name)
		return 0, fmt.Errorf("invalid storage class:%s", name)
	}

	log.Logf("class[%s] to tier[%d]\n", name, tier)
	return tier, nil
}

func sortActions(actions []*s3.Action) {
	log.Log("Sort actions.")

	sort.Slice(actions, func(i, j int) bool {
		// Expiration is sorted after transition.
		if actions[i].Name == ActionNameExpiration {
			return false
		} else if actions[j].Name == ActionNameExpiration {
			return true
		} else {
			// Transitions are sorted by tier, as default, it should be like: Tier_1 before Tier_99 before Tier_999
			if actions[i].Tier <= actions[j].Tier {
				return true
			} else {
				return false
			}
		}
	})
}

func checkValidationOfActions(actions []*s3.Action) error {
	// sort
	sortActions(actions)

	// check
	var pre *s3.Action = nil
	for _, ia := range actions {
		log.Logf("ia: %+v\n", *ia)
		if pre == nil {
			if ia.Name == ActionNameExpiration && ia.Days < ExpirationMinDays {
				// If only an expiration action for a rule, the days for that action should be more than ExpirationMinDays
				log.Logf("expiration days: %d\n", ia.Days)
				return fmt.Errorf("days for expiration must not less than %d", ExpirationMinDays)
			}
			if ia.Name == ActionNameTransition && ia.Days < TransitionMinDays {
				// If only an expiration action for a rule, the days for that action should be more than ExpirationMinDays
				log.Logf("transition days: %d\n", ia.Days)
				return fmt.Errorf("days for expiration must not less than %d", TransitionMinDays)
			}
		} else {
			if pre.Name == ActionNameExpiration {
				// Only one expiration action for each rule is supported
				return fmt.Errorf("more than one expiration action existed in one rule")
			}

			if ia.Name == ActionNameExpiration && pre.Days + ExpirationMinDays > ia.Days {
				log.Logf("pre.Days=%d, ia.Days=%d\n", pre.Days, ia.Days)
				return fmt.Errorf("object should be save in the current storage class not less than %d days before exipration(%d < %d + %d)",
					ExpirationMinDays, ia.Days, pre.Days, ExpirationMinDays)
			}

			if ia.Name == ActionNameTransition && pre.Days + LifecycleTransitionDaysStep > ia.Days {
				log.Logf("pre.Days=%d, ia.Days=%d\n", pre.Days, ia.Days)
				return fmt.Errorf("object should be save in the current storage class not less than %d days before transition(%d < %d + %d)",
					LifecycleTransitionDaysStep, ia.Days, pre.Days, LifecycleTransitionDaysStep)
			}
		}

		pre = ia
	}

	return nil
}

func (s *APIService) BucketLifecyclePut(request *restful.Request, response *restful.Response) {
	/*if !policy.Authorize(request, response, "bucket:put") {
		return
	}*/
	bucketName := request.PathParameter("bucketName")
	log.Logf("Received request for create bucket lifecycle: %s", bucketName)
	ctx := context.Background()
	bucket, _ := s.s3Client.GetBucket(ctx, &s3.Bucket{Name: bucketName})
	body := ReadBody(request)
	log.Logf("Body request is %v\n", body)
	log.Logf("%x", md5.Sum(body))
	// TODO: check the md5

	if body != nil {
		createLifecycleConf := model.LifecycleConfiguration{}
		log.Logf("Before unmarshal struct is %v\n", createLifecycleConf)
		err := xml.Unmarshal(body, &createLifecycleConf)
		log.Logf("After unmarshal struct is %v\n", createLifecycleConf)
		if err != nil {
			response.WriteError(http.StatusInternalServerError, err)
			return
		} else {
			dupIdCheck := make(map[string]interface{})
			s3RulePtrArr := make([]*s3.LifecycleRule, 0)
			for _, rule := range createLifecycleConf.Rule {
				s3Rule := s3.LifecycleRule{}

				// assign the fields

				// Check if id is duplicate
				if _, ok := dupIdCheck[rule.ID]; ok {
					log.Logf("duplication check failed, duplicate id: %s\n", rule.ID)
					response.WriteError(http.StatusBadRequest, fmt.Errorf("duplicate rule id: %s", rule.ID))
					return
				}
				// Assigning the rule ID
				dupIdCheck[rule.ID] = struct {}{}
				s3Rule.Id = rule.ID

				// Assigning the status value to s3 status
				log.Logf("Status in rule file is %v\n", rule.Status)
				s3Rule.Status = rule.Status

				// Assigning the filter, using convert function to convert xml struct to s3 struct
				s3Rule.Filter = convertRuleFilterToS3Filter(rule.Filter)

				// Create the type of action array
				actions := make([]*s3.Action, 0)
				//var actions []*s3.Action

				// loop for getting transition actions
				for _, transition := range rule.Transition {

					//Defining the Transition array and assigning the values tp populate fields
					s3Transition := s3.Action{Name: ActionNameTransition}

					//Assigning the value of days for transition to happen
					s3Transition.Days = transition.Days

					//Assigning the backend value to the s3 struct
					s3Transition.Backend = transition.Backend

					//Assigning the storage class of the object to s3 struct
					tier, err := s.class2tier(transition.StorageClass)
					if err != nil {
						response.WriteError(http.StatusBadRequest, err)
						return
					}
					s3Transition.Tier = tier

					//Adding the transition value to the main rule
					actions = append(actions, &s3Transition)
				}

				//Loop for getting the expiration actions
				for _, expiration := range rule.Expiration {
					s3Expiration := s3.Action{Name: ActionNameExpiration}
					s3Expiration.Days = expiration.Days
					actions = append(actions, &s3Expiration)
				}

				err := checkValidationOfActions(actions)
				if err != nil {
					log.Logf("check validation of actions failed: %v\n", err)
					response.WriteError(http.StatusBadRequest, err)
					return
				}

				//Assigning the actions to s3 struct
				s3Rule.Actions = actions

				s3Rule.AbortIncompleteMultipartUpload = convertRuleUploadToS3Upload(rule.AbortIncompleteMultipartUpload)
				// add to the s3 array
				s3RulePtrArr = append(s3RulePtrArr, &s3Rule)
			}
			// assign lifecycle rules to s3 bucket
			bucket.LifecycleConfiguration = s3RulePtrArr
			log.Logf("final bucket metadata is %v\n", bucket)
		}
	} else {
		log.Log("no request body provided for create lifecycle")
		response.WriteError(http.StatusBadRequest, errors.New("no request body provided"))
		return
	}

	// Create bucket with bucket name will check if the bucket exists or not, if it exists
	// it will internally call UpdateBucket function
	res, err := s.s3Client.UpdateBucket(ctx, bucket)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	log.Log("Create bucket lifecycle successfully.")
	response.WriteEntity(res)

}

func convertRuleFilterToS3Filter(filter model.Filter) *s3.LifecycleFilter {
	retFilter := s3.LifecycleFilter{}
	retFilter.Prefix = filter.Prefix
	return &retFilter
}

func convertRuleUploadToS3Upload(upload model.AbortIncompleteMultipartUpload) *s3.AbortMultipartUpload {
	retUpload := s3.AbortMultipartUpload{}
	retUpload.DaysAfterInitiation = upload.DaysAfterInitiation
	return &retUpload
}
