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
	"errors"
	"context"

	"github.com/micro/go-log"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	"github.com/opensds/multi-cloud/datamover/proto"
)

func deleteObj(objKey string, virtBucket string, bkend *BackendInfo) error {
	log.Logf("object expiration: objKey=%s, virtBucket=%s, bkend:%+v\n", objKey, virtBucket, *bkend)
	if virtBucket == "" {
		log.Logf("expiration of object[%s] is failed: virtual bucket is null.\n", objKey)
		return errors.New(DMERR_InternalError)
	}

	// delete metadata
	delMetaReq := osdss3.DeleteObjectInput{Bucket: virtBucket, Key: objKey}
	ctx := context.Background()
	_, err := s3client.DeleteObject(ctx, &delMetaReq)
	if err != nil {
		log.Logf("delete object metadata of obj[bucket:%s,objKey:%s] failed, err:%v\n",
			virtBucket, objKey, err)
		return err
	} else {
		log.Logf("Delete object metadata of obj[bucket:%s,objKey:%s] successfully.\n",
			virtBucket,	objKey)
	}

	loca := &LocationInfo{StorType:bkend.StorType, Region:bkend.Region, EndPoint:bkend.EndPoint, BucketName: bkend.BucketName,
		Access:bkend.Access, Security:bkend.Security, BakendName:bkend.BakendName, VirBucket:virtBucket}
	err = deleteObjFromBackend(objKey, loca)

	return err
}

func doExpirationAction(acReq *datamover.LifecycleActionRequest) error {
	log.Logf("Delete action: delete %s.\n", acReq.ObjKey)

	loc, err := getBackendInfo(&acReq.SourceBackend, false)
	if err != nil {
		log.Logf("exipration of %s failed because get location failed.\n", acReq.ObjKey)
	}

	err = deleteObj(acReq.ObjKey, acReq.BucketName, loc)
	if err != nil && err.Error() == DMERR_NoPermission {
		loc, err = getBackendInfo(&acReq.SourceBackend, true)
		if err != nil {
			return err
		}
		err = deleteObj(acReq.ObjKey, acReq.BucketName, loc)
	}

	return err
}

