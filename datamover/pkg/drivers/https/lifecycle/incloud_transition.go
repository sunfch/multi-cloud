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
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/datamover/proto"
	osdss3 "github.com/opensds/multi-cloud/yigs3/proto"
)

func doInCloudTransition(acReq *datamover.LifecycleActionRequest) error {
	log.Logf("in-cloud transition action: transition %s from %d to %d of %s.\n",
		acReq.ObjKey, acReq.SourceTier, acReq.TargetTier, acReq.SourceBackend)

	req := osdss3.CopyObjectRequest{
		ObjKey: acReq.ObjKey,
		SrcBucket: acReq.BucketName,
		TargetTier: acReq.TargetTier,
	}

	_, err := s3client.CopyObject(context.Background(), &req)
	if err != nil {
		log.Logf("copy object based on osds s3 failed, obj=%s, err:%v\n", acReq.ObjKey, err)
	}

	return err
}
