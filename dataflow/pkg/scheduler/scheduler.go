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

package scheduler

import (
	"context"
	"fmt"
	"os"

	"github.com/micro/go-log"
	"github.com/micro/go-micro/metadata"
	"github.com/opensds/multi-cloud/api/pkg/common"
	"github.com/opensds/multi-cloud/dataflow/pkg/db"
	"github.com/opensds/multi-cloud/dataflow/pkg/model"
	"github.com/opensds/multi-cloud/dataflow/pkg/plan"
	"github.com/opensds/multi-cloud/dataflow/pkg/scheduler/lifecycle"
	"github.com/opensds/multi-cloud/dataflow/pkg/scheduler/trigger"
	"github.com/robfig/cron"
	"strconv"
)

func LoadAllPlans() {
	ctx := metadata.NewContext(context.Background(), map[string]string{
		common.CTX_KEY_IS_ADMIN: strconv.FormatBool(true),
	})

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
			e := plan.NewPlanExecutor(&p)
			err := trigger.GetTriggerMgr().Add(ctx, &p, e)
			if err != nil {
				log.Logf("Load plan(%s) to trigger filed, %v", p.Id.Hex(), err)
				continue
			}
			log.Logf("Load plan(%s) to trigger success", p.Id.Hex())
		}
	}
}

//This scheduler will scan all buckets periodically to get lifecycle rules, and scheduling according to these rules.
func LoadLifecycleScheduler() error {
	spec := os.Getenv("LIFECYCLE_CRON_CONFIG")
	log.Logf("Value of LIFECYCLE_CRON_CONFIG is: %s\n", spec)

	//TODO: Check the validation of spec
	cn := cron.New()
	//0 */10 * * * ?
	if err := cn.AddFunc(spec, lifecycle.ScheduleLifecycle); err != nil {
		log.Logf("add lifecyecle scheduler to cron trigger failed: %v.\n", err)
		return fmt.Errorf("add lifecyecle scheduler to cron trigger failed: %v", err)
	}
	cn.Start()

	log.Log("add lifecycle scheduler to cron trigger successfully.")
	return nil
}
