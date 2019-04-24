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
	"context"
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/datamover/proto"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	mover "github.com/opensds/multi-cloud/datamover/pkg/drivers/https"
	"strconv"
	"os"
	"fmt"
)
var PART_SIZE int64 = 16 * 1024 * 1024 //The max object size that can be moved directly, default is 16M.

func move(ctx context.Context, obj *osdss3.Object, src *BackendInfo, dest *BackendInfo) error {
	//move object
	part_size, err := strconv.ParseInt(os.Getenv("PARTSIZE"), 10, 64)
	log.Logf("part_size=%d, err=%v.\n", part_size, err)
	if err == nil {
		//part_size must be more than 5M and less than 100M
		if part_size >= 5 && part_size <= 100 {
			PART_SIZE = part_size * 1024 * 1024
			log.Logf("Set PART_SIZE to be %d.\n", PART_SIZE)
		}
	}

	srcLoc := &LocationInfo{StorType:src.StorType, Region:src.Region, EndPoint:src.EndPoint, BucketName: obj.BucketName,
		Access:src.Access, Security:src.Security, BakendName:src.BakendName}
	targetLoc := &LocationInfo{StorType:dest.StorType, Region:dest.Region, EndPoint:dest.EndPoint, BucketName: obj.BucketName,
		Access:dest.Access, Security:dest.Security, BakendName:dest.BakendName}
	if obj.Size <= PART_SIZE {
		err = mover.MoveObj(obj, srcLoc, targetLoc)
	} else {
		err = mover.MultipartMoveObj(obj, srcLoc, targetLoc)
	}

	return err
}

func doCrossCloudTransition(acReq *datamover.LifecycleActionRequest) error {
	log.Logf("Cross-cloud transition action: transition %s from %d of %s to %d of %s.\n",
		acReq.ObjKey, acReq.SourceTier, acReq.SourceBackend, acReq.TargetTier, acReq.TargetBackend)

	src, err := getBackendInfo(&acReq.BucketName, &acReq.SourceBackend, false)
	if err != nil {
		log.Logf("Cross-cloud transition of %s failed:%v\n", acReq.ObjKey, err)
		return err
	}
	target, err := getBackendInfo(&acReq.BucketName, &acReq.TargetBackend, false)
	if err != nil {
		log.Logf("Cross-cloud transition of %s failed:%v\n", acReq.ObjKey, err)
		return err
	}

	obj := osdss3.Object{ObjectKey:acReq.ObjKey, Size:acReq.ObjSize}
	err = move(context.Background(), &obj, src, target)

	if err != nil {
		log.Logf("Cross-cloud transition of %s failed:%v\n", acReq.ObjKey, err)
		return err
	}

	//update meta data.
	setting := make(map[string]string)
	setting[OBJMETA_TIER] = fmt.Sprintf("%s", acReq.TargetTier)
	setting[OBJMETA_BACKEND] = acReq.TargetBackend
	//req := osdss3.UpdateObjMetaRequest{acReq.ObjKey, acReq.BucketName}
	req := osdss3.UpdateObjMetaRequest{ObjKey:acReq.ObjKey, BucketName:acReq.BucketName, Setting:setting}
	_, err = s3client.UpdateObjMeta(context.Background(), &req)
	if err != nil {
		log.Logf("Update tier of object[%s] to %d failed:%v.\n", acReq.ObjKey, acReq.TargetTier, err)
		//TODO: Log it and let another module to handle.
	}

	return nil
}
