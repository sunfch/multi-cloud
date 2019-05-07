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
	"strconv"
	"os"
	"fmt"

	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/datamover/proto"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	mover "github.com/opensds/multi-cloud/datamover/pkg/drivers/https"
)

//The max object size that can be moved directly, default is 16M.
var PART_SIZE int64 = 16 * 1024 * 1024

func copyObj(ctx context.Context, obj *osdss3.Object, src *BackendInfo, dest *BackendInfo, className *string) error {
	// move object
	part_size, err := strconv.ParseInt(os.Getenv("PARTSIZE"), 10, 64)
	log.Logf("part_size=%d, err=%v.\n", part_size, err)
	if err == nil {
		// part_size must be more than 5M and less than 100M
		if part_size >= 5 && part_size <= 100 {
			PART_SIZE = part_size * 1024 * 1024
			log.Logf("Set PART_SIZE to be %d.\n", PART_SIZE)
		}
	}

	srcLoc := &LocationInfo{StorType:src.StorType, Region:src.Region, EndPoint:src.EndPoint, BucketName:src.BucketName,
		Access:src.Access, Security:src.Security, BakendName:src.BakendName, VirBucket:obj.BucketName}
	targetLoc := &LocationInfo{StorType:dest.StorType, Region:dest.Region, EndPoint:dest.EndPoint, BucketName:dest.BucketName,
		Access:dest.Access, Security:dest.Security, BakendName:dest.BakendName, ClassName:*className, VirBucket:obj.BucketName}
	if obj.Size <= PART_SIZE {
		err = mover.MoveObj(obj, srcLoc, targetLoc)
	} else {
		err = mover.MultipartMoveObj(obj, srcLoc, targetLoc)
	}

	// TODO: Need to confirm the integrity by comparing Etags.
	return err
}

func doCrossCloudTransition(acReq *datamover.LifecycleActionRequest) error {
	log.Logf("Cross-cloud transition action: transition %s from %d of %s to %d of %s.\n",
		acReq.ObjKey, acReq.SourceTier, acReq.SourceBackend, acReq.TargetTier, acReq.TargetBackend)

	src, err := getBackendInfo(&acReq.SourceBackend, false)
	if err != nil {
		log.Logf("cross-cloud transition of %s failed:%v\n", acReq.ObjKey, err)
		return err
	}
	target, err := getBackendInfo(&acReq.TargetBackend, false)
	if err != nil {
		log.Logf("cross-cloud transition of %s failed:%v\n", acReq.ObjKey, err)
		return err
	}

	className, err := getStorageClassName(acReq.TargetTier, target.StorType)
	if err != nil {
		log.Logf("cross-cloud transition of %s failed because target tier is not supported.\n", acReq.ObjKey)
		return err
	}

	log.Logf("move object[%s] from [%+v] to [%+v]\n", acReq.ObjKey, src, target)
	obj := osdss3.Object{ObjectKey:acReq.ObjKey, Size:acReq.ObjSize, BucketName:acReq.BucketName}
	err = copyObj(context.Background(), &obj, src, target, &className)
	if err != nil && err.Error() == DMERR_NoPermission{
		log.Logf("cross-cloud transition of %s failed:%v\n", acReq.ObjKey, err)
		// In case credentials is changed.
		src, _ = getBackendInfo(&acReq.SourceBackend, true)
		target, _ = getBackendInfo(&acReq.TargetBackend, true)
		err = copyObj(context.Background(), &obj, src, target, &className)
	}
	if err != nil {
		log.Logf("cross-cloud transition of %s failed:%v\n", acReq.ObjKey, err)
		return err
	}

	// update meta data.
	setting := make(map[string]string)
	setting[OBJMETA_TIER] = fmt.Sprintf("%d", acReq.TargetTier)
	setting[OBJMETA_BACKEND] = acReq.TargetBackend
	req := osdss3.UpdateObjMetaRequest{ObjKey:acReq.ObjKey, BucketName:acReq.BucketName, Setting:setting, LastModified:acReq.LastModified}
	_, err = s3client.UpdateObjMeta(context.Background(), &req)
	var loca *LocationInfo
	if err != nil {
		// if update metadata failed, then delete object from target storage backend.
		loca = &LocationInfo{StorType:target.StorType, Region:target.Region, EndPoint:target.EndPoint, BucketName:target.BucketName,
			Access:target.Access, Security:target.Security, BakendName:target.BakendName, VirBucket:obj.BucketName}
	} else {
		// if update metadata successfully, then delete object from source storage backend.
		loca = &LocationInfo{StorType:src.StorType, Region:src.Region, EndPoint:src.EndPoint, BucketName:src.BucketName,
			Access:src.Access, Security:src.Security, BakendName:src.BakendName, VirBucket:obj.BucketName}
	}

	// delete object from the storage backend.
	deleteObjFromBackend(acReq.ObjKey, loca)

	return nil
}
