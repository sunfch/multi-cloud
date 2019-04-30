package lifecycle

import (
	"github.com/opensds/multi-cloud/datamover/proto"
	"encoding/json"
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/dataflow/pkg/kafka"
	"github.com/opensds/multi-cloud/s3/proto"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	. "github.com/opensds/multi-cloud/dataflow/pkg/utils"
	. "github.com/opensds/multi-cloud/dataflow/pkg/model"
	"github.com/opensds/multi-cloud/dataflow/pkg/db"
	"sort"
	"fmt"
	"sync"
	"strconv"
)

var topicLifecycle = "lifecycle"
var s3client = osdss3.NewS3Service("s3", client.DefaultClient)
type InterRules []*InternalLifecycleRule
//type Int2String map[int32]string
// map from cloud vendor name to it's map relation relationship between internal tier to it's storage class name.
//var Int2ExtTierMap map[string]*Int2String
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

//Get liecycle rules for each bucket from db, and schedule according to those rules.
func ScheduleLifecycle() {
	log.Log("[ScheduleLifecycle] begin ...")
	//Load transition map.
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

	//Get bucket list.
	listReq := s3.BaseRequest{Id: "test"}
	listRsp, err := s3client.ListBuckets(context.Background(), &listReq)
	if err != nil {
		log.Logf("[ScheduleLifecycle]List buckets failed: %v.\n", err)
		return
	}

	for _, v := range listRsp.Buckets {
		//For each bucket, get the lifecycle rules, and schedule each rule.
		if v.LifecycleConfiguration == nil {
			log.Logf("[ScheduleLifecycle]Bucket[%s] has no lifecycle rule.\n", v.Name)
			continue
		}

		log.Logf("[ScheduleLifecycle]Bucket[%s] has lifecycle rule.\n", v.Name)

		err := handleBucketLifecyle(v.Name, v.LifecycleConfiguration)
		if err != nil {
			log.Logf("[ScheduleLifecycle]Handle bucket lifecycle for bucket[%s] failed, err:%v.\n", v.Name, err)
			continue
		}
	}

	log.Log("[ScheduleLifecycle] end ...")
}

//Need to lock the bucket, incase the schedule period is too shoort and the bucket is scheduled at the same time.
//Need to consider confliction between rules.
func handleBucketLifecyle(bucket string, rules []*osdss3.LifecycleRule) error {
	// Translate rules set by user to internal rules which can be sorted.

	//Lifecycle scheduling must be mutual excluded among several schedulers, so get lock first.
	ret := db.DbAdapter.LockBucketLifecycleSched(bucket)
	for i := 0; i < 3; i++ {
		if ret == LockSuccess {
			break
		} else if ret == LockBusy {
			return fmt.Errorf("Lifecycle scheduleing of bucket[%s] is in progress.")
		} else {
			//Try to lock again, try three times at most
			ret = db.DbAdapter.LockBucketLifecycleSched(bucket)
		}
	}
	if ret != LockSuccess {
		log.Logf("Lock scheduling failed.\n")
		return fmt.Errorf("Internal error: Lock scheduling failed.")
	}
	//Make sure unlock before return
	defer db.DbAdapter.UnlockBucketLifecycleSched(bucket)

	var inRules InterRules
	for _, rule := range rules {
		if rule.Status == RuleStatusOff {
			continue
		}

		for _, ac := range rule.Actions {
			var acType int
			if ac.Name == ActionNameExpiration {//Eexpiration
				acType = ActionExpiration
			} else if ac.Backend == "" {//ac.Name == "Transition" and no backend specified.
				acType = ActionIncloudTransition
			} else {  //ac.Name == "Transition" and backend is specified.
				acType = ActionCrosscloudTransition
			}

			v := InternalLifecycleRule{Bucket: bucket, Filter:InternalLifecycleFilter{Prefix:rule.Filter.Prefix},
			Days:ac.Days, ActionType:acType, Tier: ac.Tier, Backend: ac.Backend}
			inRules = append(inRules, &v)
		}
		//objs := listObjsWithFilter(bucket, rule.Filter)
	}

	//Sort rules, delete actions has the highest priority, in-cloud transition has the second priority,
	// while cross-cloud has the lowest priority.
	sort.Sort(inRules)
	// Begin: For debug
	for _, v := range inRules {
		log.Logf("rule: %+v\n", *v)
	}
	// End: For debug

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
	//Get objects by communicating with s3 service.
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
		Limit: limit,
	}
	ctx := context.Background()
	s3rsp, err := s3client.ListObjects(ctx, &s3req)
	if err != nil {
		log.Logf("List objects failed, req: %v.\n", s3req)
		return nil, err
	}

	return s3rsp.ListObjects, nil
}

func schedAccordingSortedRules (inRules *InterRules) error {
	dupCheck := map[string]struct{}{}
	for _, r := range *inRules {
		var offset, limit int32 = 0, 1000
		for {
			objs, err := getObjects(r, offset, limit)
			if err != nil {
				break
			}
			num := int32(len(objs))
			offset += num
			//Check if the object exist in the dupCheck map.
			for _, obj := range objs {
				log.Logf("***obj:%s\n", obj.ObjectKey)
				if obj.IsDeleteMarker == "1" {
					log.Logf("DeleteMarker of object[%s] is set, action is cancelled.\n", obj.ObjectKey)
					continue
				}
				if _, ok := dupCheck[obj.ObjectKey]; !ok {
					//If not exist, then send request to datamover through kafka and add object key to dupCheck.
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
						log.Logf("Transition object[%s] from tier[%d] to tier[%d] is invalid.\n", obj.ObjectKey, obj.Tier, r.Tier)
						continue
					}
					log.Logf("Lifecycle action: object=[%s] type=%[%s] tier[%d] to tier[%d].\n",
						obj.ObjectKey, r.ActionType, obj.Tier, r.Tier)
					acreq := datamover.LifecycleActionRequest{
						ObjKey:obj.ObjectKey,
						BucketName:obj.BucketName,
						Action:action,
						SourceTier:obj.Tier,
						TargetTier:r.Tier,
						SourceBackend:obj.Backend,
						TargetBackend:r.Backend,
					}

					//If send failed, then ignore it, because it will be re-sent in the next period.
					sendActionRequest(&acreq)

					//Add object key to dupCheck.
					dupCheck[obj.ObjectKey] = struct{}{}
					log.Logf("dupCheck:%v\n", dupCheck)
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
		log.Logf("Marshal run job request failed, err:%v\n", data)
		return err
	}


	return kafka.ProduceMsg(topicLifecycle, data)
}

func (r InterRules) Len() int {
	return len(r)
}

// Less reports whether the element with index i should sort before the element with index j.
func (r InterRules) Less(i, j int) bool {
	var ret bool
	if r[i].ActionType < r[j].ActionType {
		ret = true
	} else if r[i].ActionType > r[j].ActionType {
		ret = false
	} else {
		if r[i].Days <= r[j].Days {
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