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

package lifecycle

import (
	"errors"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/datamover/proto"
	"encoding/json"
	"github.com/opensds/multi-cloud/dataflow/pkg/utils"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	"github.com/opensds/multi-cloud/datamover/pkg/db"
	"github.com/micro/go-micro/client"
	"github.com/opensds/multi-cloud/backend/proto"
	"sync"
)

var bkendInfo map[string]*BackendInfo
var s3client osdss3.S3Service
var bkendclient backend.BackendService
var mutex sync.RWMutex
type Int2String map[int32]string
// map from cloud vendor name to it's map relation relationship between internal tier to it's storage class name.
var Int2ExtTierMap map[string]*Int2String

func Init() {
	log.Logf("Lifecycle datamover init.")
	s3client = osdss3.NewS3Service("s3", client.DefaultClient)
	bkendclient = backend.NewBackendService("backend", client.DefaultClient)
	bkendInfo = make(map[string]*BackendInfo)
}

func HandleMsg(msgData []byte) error {
	var acReq datamover.LifecycleActionRequest
	err := json.Unmarshal(msgData, &acReq)
	if err != nil {
		log.Logf("Unmarshal lifecycle action request failed, err:%v\n", err)
		return err
	}

	go doAction(&acReq)

	return nil
}

func doAction(acReq *datamover.LifecycleActionRequest){
	acType := int(acReq.Action)
	switch acType {
	case utils.ActionCrosscloudTransition:
		doCrossCloudTransition(acReq)
	case utils.ActionIncloudTransition:
		doInCloudTransition(acReq)
	case utils.ActionExpiration:
		doExpirationAction(acReq)
	default:
		log.Logf("Unsupported action type: %d.\n", acType)
	}
}

//force means get location information from database not from cache.
func getBackendInfo(bucketName *string, backendName *string, force bool) (*BackendInfo, error) {
	if !force {
		loc, exist := bkendInfo[*backendName]
		if exist {
			//loc.VirBucket = *bucketName
			return loc, nil
		}
	}

	if *backendName == "" {
		log.Log("Get backend information failed, backendName is null.\n")
		return nil, errors.New(DMERR_InternalError)
	}

	bk, err := db.DbAdapter.GetBackendByName(*backendName)
	if err != nil {
		log.Logf("Get backend[%s] information failed, err:%v\n", backendName, err)
		return nil, err
	} else {
		loca := &BackendInfo{bk.Type, bk.Region, bk.Endpoint, bk.BucketName,
		bk.Access, bk.Security, *backendName}
		log.Logf("Refresh backend[name:%s, loca:%+v] successfully.\n", *backendName, *loca)
		bkendInfo[*backendName] = loca
		return loca, nil
	}
}