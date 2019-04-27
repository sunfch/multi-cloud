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
	"strconv"
	"sync"
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
	if len(res.TransitionMap) == 0 {
		log.Log("Get transition map failed")
		return fmt.Errorf("Get tier definition failed")
	}

	/*Int2ExtTierMap = make(map[string]*Int2String)
	for k, v := range res.Tier2Name {
		val := make(Int2String)
		for k1, v1 := range v.Lst {
			val[k1] = v1
		}
		Int2ExtTierMap[k] = &val
	}*/

	TransitionMap = make(map[string]struct{})
	for k, v := range res.TransitionMap {
		l := v.Tier
		for _, t := range l {
			s := fmt.Sprintf("%s:%s", k, t)
			TransitionMap[s] = struct{}{}
		}
	}

	return nil
}

//Get liecycle rules for each bucket from db, and schedule according to those rules.
func ScheduleLifecycle() {
	//Load transition map.
	{
		mutext.Lock()
		defer mutext.Unlock()
		if len(TransitionMap) == 0 {
			err := loadStorageClassDefinition()
			if err != nil {
				log.Logf("Load transition map faild: %v.\n", err)
				return
			}
		}
	}

	//Get bucket list.
	//TODO: Need to consider paging query and setting create time as index on collection of object.
	//s3client := osdss3.NewS3Service("s3", client.DefaultClient)
	//TODO: when list buckets, id of request shouled not be necessary, need to change the s3 api.
	listReq := s3.BaseRequest{Id: "test"}
	listRsp, err := s3client.ListBuckets(context.Background(), &listReq)
	if err != nil {
		log.Logf("List buckets faild: %v.\n", err)
		return
	}

	for _, v := range listRsp.Buckets {
		//For each bucket, get the lifecycle rules, and schedule each rule.
		if v.LifecycleConfiguration == nil {
			log.Logf("Bucket[%s] has no lifecycle rule.\n", v.Name)
			continue
		}

		err := handleBucketLifecyle(v.Name, v.LifecycleConfiguration)
		if err != nil {
			log.Logf("Handle bucket lifecycle for bucket[%s] failed, err:%v.\n", v.Name, err)
			continue
		}
	}
}

//Need to lock the bucket, incase the schedule period is too shoort and the bucket is scheduled at the same time.
//Need to consider confliction between rules.
func handleBucketLifecyle(bucket string, rules []*osdss3.LifecycleRule) error {
	// Translate rules set by user to internal rules which can be sorted.

	//Lifecycle scheduling must be mutual excluded among several schedulers, so get lock first.
	ret := db.DbAdapter.LockBucketLifecycleSched(bucket)
	for i := 0; i < 3; i++ {
		if ret == LockSuccess {
			//Make sure unlock before return
			defer db.DbAdapter.UnlockBucketLifecycleSched(bucket)
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

	var inRules InterRules
	for _, rule := range rules {
		if rule.Status == "off" {
			continue
		}

		for _, ac := range rule.Actions {
			var acType int
			if ac.Name == "Expiration" {//Eexpiration
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
		log.Logf("rule: %v\n", *v)
	}
	// End: For debug

	err := schedAccordingSortedRules(&inRules)

	return err
}

func checkTransitionValidation(source int32, destination int32) bool {
	valid := true
	key := fmt.Sprintf("%s:%s", source, destination)
	if _, ok := TransitionMap[key]; !ok {
		valid = false
	}

	return valid
}

func schedAccordingSortedRules (inRules *InterRules) error {
	dupCheck := map[string]struct{}{}
	for _, r := range *inRules {
		//List object by communicating with s3 service.
		filt := make(map[string]string)
		filt[KObjKey] = "^" + r.Filter.Prefix
		days := fmt.Sprintf("%s", r.Days)
		filt[KLastModified] = days
		filt[KStorageTier] = strconv.Itoa(int(r.Tier))
		s3req := osdss3.ListObjectsRequest{
			Bucket:r.Bucket,
			Filter:filt,
		}
		ctx := context.Background()
		s3rsp, err := s3client.ListObjects(ctx, &s3req)
		if err != nil {
			log.Logf("List objects failed, req: %v.\n", s3req)
			continue
		}
		//Check if the object exist in the dupCheck map.
		for _, obj := range s3rsp.ListObjects {
			if _, ok := dupCheck[obj.ObjectKey]; !ok {
				//If not exist, then send request to datamover through kafka and add object key to dupCheck.
				if r.ActionType != ActionExpiration && obj.Backend == r.Backend && obj.Tier == r.Tier {
					// For transition, if target backend and storage class is the same as source backend and storage class, then no transition is need.
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
					continue
				}

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
			}
		}
	}

	return nil
}

func sendActionRequest(req *datamover.LifecycleActionRequest) error {
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
	if r[i].Days < r[j].Days {
		ret = true
	} else if (r[i].Days > r[j].Days) {
		ret = false
	} else { // This means r[i].Days == r[j].Days
		if (r[i].ActionType < r[j].ActionType) {
			ret = true
		} else if (r[i].ActionType > r[j].ActionType) {
			ret = false
		} else { // This means r[i].ActionType == r[j].ActionType
		    //TODO:??
		    ret = true
		}
	}

	return ret
}

func (r InterRules) Swap(i, j int) {
	tmp := r[i]
	r[i] = r[j]
	r[j] = tmp
}