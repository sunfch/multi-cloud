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

package lifecycle

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"strconv"

	"github.com/micro/go-log"
	"github.com/micro/go-micro/client"
	"github.com/opensds/multi-cloud/dataflow/pkg/db"
	"github.com/opensds/multi-cloud/dataflow/pkg/kafka"
	. "github.com/opensds/multi-cloud/dataflow/pkg/model"
	. "github.com/opensds/multi-cloud/dataflow/pkg/utils"
	datamover "github.com/opensds/multi-cloud/datamover/proto"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	s3 "github.com/opensds/multi-cloud/s3/proto"
	"golang.org/x/net/context"
)

var topicLifecycle = "lifecycle"
var s3client = osdss3.NewS3Service("s3", client.DefaultClient)

type InterRules []*InternalLifecycleRule

// map from a specific tier to an array of tiers, that means transition can happens from the specific tier to those tiers in the array.
var TransitionMap map[string]struct{}
var mutext sync.Mutex

func loadStorageClassDefinition() error {
	res, _ := s3client.GetTierMap(context.Background(), &s3.BaseRequest{})
	if len(res.Transition) == 0 {
		log.Log("Get transition map failed")
		return fmt.Errorf("Get tier definition failed")
	} else {
		log.Logf("res.Transition:%v", res.Transition)
		log.Logf("res.Tier2Name:%+v", res.Tier2Name)
	}

	TransitionMap = make(map[string]struct{})
	for _, v := range res.Transition {
		TransitionMap[v] = struct{}{}
	}

	return nil
}

// Get liecycle rules for each bucket from db, and schedule according to those rules.
func ScheduleLifecycle() {
	log.Log("[ScheduleLifecycle] begin ...")
	// Load transition map.
	{
		mutext.Lock()
		defer mutext.Unlock()
		if len(TransitionMap) == 0 {
			err := loadStorageClassDefinition()
			if err != nil {
				log.Logf("[ScheduleLifecycle]Load storage classes failed: %v.\n", err)
				return
			}
		}
	}

	// Get bucket list.
	listReq := s3.BaseRequest{Id: "test"}
	listRsp, err := s3client.ListBuckets(context.Background(), &listReq)
	if err != nil {
		log.Logf("[ScheduleLifecycle]List buckets failed: %v.\n", err)
		return
	}

	for _, v := range listRsp.Buckets {
		// For each bucket, get the lifecycle rules, and schedule each rule.
		if v.LifecycleConfiguration == nil {
			log.Logf("[ScheduleLifecycle]Bucket[%s] has no lifecycle rule.\n", v.Name)
			continue
		}

		log.Logf("[ScheduleLifecycle]Bucket[%s] has lifecycle rule.\n", v.Name)

		err := handleBucketLifecyle(v.Name, v.Backend, v.LifecycleConfiguration)
		if err != nil {
			log.Logf("[ScheduleLifecycle]Handle bucket lifecycle for bucket[%s] failed, err:%v.\n", v.Name, err)
			continue
		}
	}

	log.Log("[ScheduleLifecycle] end ...")
}

// Need to lock the bucket, incase the schedule period is too shoort and the bucket is scheduled at the same time.
// Need to consider confliction between rules.
func handleBucketLifecyle(bucket string, defaultBackend string, rules []*osdss3.LifecycleRule) error {
	// Translate rules set by user to internal rules which can be sorted.

	// Lifecycle scheduling must be mutual excluded among several schedulers, so get lock first.
	ret := db.DbAdapter.LockBucketLifecycleSched(bucket)
	for i := 0; i < 3; i++ {
		if ret == LockSuccess {
			break
		} else if ret == LockBusy {
			return fmt.Errorf("Lifecycle scheduleing of bucket[%s] is in progress.")
		} else {
			// Try to lock again, try three times at most
			ret = db.DbAdapter.LockBucketLifecycleSched(bucket)
		}
	}
	if ret != LockSuccess {
		log.Logf("Lock scheduling failed.\n")
		return fmt.Errorf("Internal error: Lock scheduling failed.")
	}
	// Make sure unlock before return
	defer db.DbAdapter.UnlockBucketLifecycleSched(bucket)

	var inRules InterRules
	for _, rule := range rules {
		if rule.Status == RuleStatusDisabled {
			continue
		}

		for _, ac := range rule.Actions {
			if ac.Backend == "" {
				// if no backend specified, then use the default backend of the bucket
				ac.Backend = defaultBackend
			}
			var acType int
			if ac.Name == ActionNameExpiration {
				// Eexpiration
				acType = ActionExpiration
			} else if ac.Backend == defaultBackend {
				acType = ActionIncloudTransition
			} else {
				acType = ActionCrosscloudTransition
			}

			v := InternalLifecycleRule{Bucket: bucket, Days: ac.GetDays(), ActionType: acType, Tier: ac.GetTier(), Backend: ac.GetBackend()}
			if rule.GetFilter() != nil {
				v.Filter = InternalLifecycleFilter{Prefix: rule.Filter.Prefix}
			}
			inRules = append(inRules, &v)
		}
	}

	// Sort rules, expiration actions has the highest priority, for the same action, bigger Days has higher priority.
	sort.Sort(inRules)
	// Begin: Log for debug
	for _, v := range inRules {
		log.Logf("rule: %+v\n", *v)
	}
	// End: Log for debug

	err := schedAccordingSortedRules(&inRules)

	return err
}

func checkTransitionValidation(source int32, destination int32) bool {
	valid := true
	key := fmt.Sprintf("%d:%d", source, destination)
	if _, ok := TransitionMap[key]; !ok {
		valid = false
	}

	return valid
}

func getObjects(r *InternalLifecycleRule, offset, limit int32) ([]*osdss3.Object, error) {
	// Get objects by communicating with s3 service.
	filt := make(map[string]string)
	filt[KObjKey] = "^" + r.Filter.Prefix
	modifyFilt := fmt.Sprintf("{\"gte\":\"%d\"}", r.Days)
	filt[KLastModified] = modifyFilt
	if r.ActionType != ActionExpiration {
		filt[KStorageTier] = strconv.Itoa(int(r.Tier))
	}

	s3req := osdss3.ListObjectsRequest{
		Bucket: r.Bucket,
		Filter: filt,
		Offset: offset,
		Limit:  limit,
	}
	ctx := context.Background()
	log.Logf("ListObjectsRequest:%+v\n", s3req)
	s3rsp, err := s3client.ListObjects(ctx, &s3req)
	if err != nil {
		log.Logf("list objects failed, req: %v.\n", s3req)
		return nil, err
	}

	return s3rsp.ListObjects, nil
}

func schedAccordingSortedRules(inRules *InterRules) error {
	dupCheck := map[string]struct{}{}
	for _, r := range *inRules {
		//log.Logf("r:%v\n", *r)
		var offset, limit int32 = 0, 1000
		for {
			objs, err := getObjects(r, offset, limit)
			if err != nil {
				break
			}
			num := int32(len(objs))
			offset += num
			//log.Logf("offset=%d, num=%d\n", offset, num)
			// Check if the object exist in the dupCheck map.
			for _, obj := range objs {
				if obj.IsDeleteMarker == "1" {
					log.Logf("DeleteMarker of object[%s] is set, no lifecycle action need.\n", obj.ObjectKey)
					continue
				}
				if _, ok := dupCheck[obj.ObjectKey]; !ok {
					// Not exist means this object has is not processed in this round of scheduling.
					if r.ActionType != ActionExpiration && obj.Backend == r.Backend && obj.Tier == r.Tier {
						// For transition, if target backend and storage class is the same as source backend and storage class, then no transition is need.
						log.Logf("No need transition for object[%s], backend=%s, tier=%d\n", obj.ObjectKey, r.Backend, r.Tier)
						continue
					}

					//Send request.
					var action int32
					if r.ActionType == ActionExpiration {
						action = int32(ActionExpiration)
					} else if obj.Backend == r.Backend {
						action = int32(ActionIncloudTransition)
					} else {
						action = int32(ActionCrosscloudTransition)
					}

					if r.ActionType != ActionExpiration && checkTransitionValidation(obj.Tier, r.Tier) != true {
						log.Logf("transition object[%s] from tier[%d] to tier[%d] is invalid.\n", obj.ObjectKey, obj.Tier, r.Tier)
						continue
					}
					log.Logf("Lifecycle action: object=[%s] type=[%d] source-tier=[%d] target-tier=[%d] source-backend=[%s] target-backend=[%s].\n",
						obj.ObjectKey, r.ActionType, obj.Tier, r.Tier, obj.Backend, r.Backend)
					acreq := datamover.LifecycleActionRequest{
						ObjKey:        obj.ObjectKey,
						BucketName:    obj.BucketName,
						Action:        action,
						SourceTier:    obj.Tier,
						TargetTier:    r.Tier,
						SourceBackend: obj.Backend,
						TargetBackend: r.Backend,
						ObjSize:       obj.Size,
						LastModified:  obj.LastModified,
					}

					// If send failed, then ignore it, because it will be re-sent in the next period.
					sendActionRequest(&acreq)

					// Add object key to dupCheck so it will not be processed repeatedly in this round or scheduling.
					dupCheck[obj.ObjectKey] = struct{}{}
				} else {
					log.Logf("Object[%s] is already handled in this schedule time.\n", obj.ObjectKey)
				}
			}
			if num < limit {
				break
			}
		}
	}

	return nil
}

func sendActionRequest(req *datamover.LifecycleActionRequest) error {
	log.Logf("Send lifecycle request to datamover: %v\n", req)
	data, err := json.Marshal(*req)
	if err != nil {
		log.Logf("marshal run job request failed, err:%v\n", data)
		return err
	}

	return kafka.ProduceMsg(topicLifecycle, data)
}

func (r InterRules) Len() int {
	return len(r)
}

// Less reports whether the element with index i should sort before the element with index j.
// Expiration action has higher priority than transition action.
// For the same action type, bigger Days has higher priority.
func (r InterRules) Less(i, j int) bool {
	var ret bool
	if r[i].ActionType < r[j].ActionType {
		ret = true
	} else {
		if r[i].Days >= r[j].Days {
			ret = true
		} else {
			ret = false
		}
	}

	return ret
}

func (r InterRules) Swap(i, j int) {
	tmp := r[i]
	r[i] = r[j]
	r[j] = tmp
}
