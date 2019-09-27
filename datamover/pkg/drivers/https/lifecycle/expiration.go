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
	log "github.com/sirupsen/logrus"
	"github.com/opensds/multi-cloud/datamover/proto"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
)

func doExpirationAction(acReq *datamover.LifecycleActionRequest) error {
	log.Infof("delete action: delete %s.\n", acReq.ObjKey)

	req := osdss3.DeleteObjectInput{Bucket: acReq.BucketName, Key: acReq.ObjKey}
	ctx := context.Background()
	_, err := s3client.DeleteObject(ctx, &req)
	if err != nil {
		// if it is deleted failed, it will be delete again in the next schedule round
		log.Infof("delete object[bucket:%s,objKey:%s] failed, err:%v\n",
			acReq.BucketName, acReq.ObjKey, err)
		return err
	} else {
		log.Infof("delete object[bucket:%s,objKey:%s] successfully.\n",
			acReq.BucketName, acReq.ObjKey)
	}

	return err
}
