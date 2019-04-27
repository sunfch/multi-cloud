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

package scheduler

import (
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/api/pkg/filters/context"
	"github.com/opensds/multi-cloud/dataflow/pkg/db"
	"github.com/opensds/multi-cloud/dataflow/pkg/model"
	"github.com/opensds/multi-cloud/dataflow/pkg/plan"
	"github.com/opensds/multi-cloud/dataflow/pkg/scheduler/trigger"
	"github.com/opensds/multi-cloud/dataflow/pkg/scheduler/lifecycle"
	"fmt"
	"github.com/robfig/cron"
)

var DEFAULT_LIFECYLE_SCHED_PERIOD = 1440

func LoadAllPlans() {
	ctx := context.NewAdminContext()

	offset := model.DefaultOffset
	limit := model.DefaultLimit

	planNum := 0
	for offset == 0 || planNum > 0 {
		plans, err := db.DbAdapter.ListPlan(ctx, limit, offset, nil)
		if err != nil {
			log.Logf("Get all plan faild, %v", err)
			break
		}
		planNum = len(plans)
		if planNum == 0 {
			break
		} else {
			offset += planNum
		}
		for _, p := range plans {
			if p.PolicyId == "" || !p.PolicyEnabled {
				continue
			}
			e := plan.NewPlanExecutor(ctx, &p)
			err := trigger.GetTriggerMgr().Add(ctx, &p, e)
			if err != nil {
				log.Logf("Load plan(%s) to trigger filed, %v", p.Id.Hex(), err)
				continue
			}
			log.Logf("Load plan(%s) to trigger success", p.Id.Hex())
		}
	}
}

//This scheduler will traversing all buckets periodically to get lifecycle rules, and scheduling according to these rules.
func LoadLifecycleScheduler() error {
	/*period, err := strconv.ParseInt(os.Getenv("LIFECYCLE_SCHEDULE_PERIOD"), 10, 64)
	log.Logf("Value of LIFECYCLE_SCHEDULE_PERIOD is: %d.\n", period)
	//period should not be less than 60 minutes or more than 1 week.
	if err != nil || period < 60 || period > 10080 {
		period = int64(DEFAULT_LIFECYLE_SCHED_PERIOD)
	}
	log.Logf("Lifecycle scheduler period is: %d minutes.\n", period)
	*/

	cn := cron.New()
	//Scheduler at 00:00:00 ever day.
	if err := cn.AddFunc("0 0 0 * * ?", lifecycle.ScheduleLifecycle); err != nil {
		log.Logf("Add lifecyecle scheduler to cron trigger failed: %v.\n", err)
		return fmt.Errorf("Add lifecyecle scheduler to cron trigger failed: %v", err)
	}

	log.Log("Add lifecyecle scheduler to cron trigger successfully.")
	return nil
}
