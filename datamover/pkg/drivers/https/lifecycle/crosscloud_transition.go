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
	"context"
	"errors"
	"sync"

	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/datamover/proto"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
)

// If transition for an object is in-progress, then the next transition message will be abandoned.
var InProgressObjs map[string]struct{}

func doCrossCloudTransition(acReq *datamover.LifecycleActionRequest) error {
	log.Logf("cross-cloud transition action: transition %s from %d of %s to %d of %s.\n",
		acReq.ObjKey, acReq.SourceTier, acReq.SourceBackend, acReq.TargetTier, acReq.TargetBackend)

	// add object to InProgressObjs
	if InProgressObjs == nil {
		var mutex sync.Mutex
		mutex.Lock()
		if InProgressObjs == nil {
			InProgressObjs = make(map[string]struct{})
		}
		mutex.Unlock()
	}

	if _, ok := InProgressObjs[acReq.ObjKey]; !ok {
		InProgressObjs[acReq.ObjKey] = struct{}{}
	} else {
		log.Logf("the transition of object[%s] is in-progress\n", acReq.ObjKey)
		return errors.New(DMERR_TransitionInprogress)
	}

	req := osdss3.CopyObjectRequest{
		ObjKey: acReq.ObjKey,
		SrcBucket: acReq.BucketName,
		TargetBackend: acReq.TargetBackend,
		TargetTier: acReq.TargetTier,
	}

	_, err := s3client.CopyObject(context.Background(), &req)
	if err != nil {
		log.Logf("copy object based on osds s3 failed, obj=%s, err:%v\n", acReq.ObjKey, err)
	}

	// remove object from InProgressObjs
	delete(InProgressObjs, acReq.ObjKey)

	return err
}
