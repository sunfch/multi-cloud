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
	"fmt"
	"errors"
	"context"
	"github.com/micro/go-log"
	"github.com/opensds/multi-cloud/datamover/proto"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	flowtype "github.com/opensds/multi-cloud/dataflow/pkg/model"
	"github.com/opensds/multi-cloud/datamover/pkg/amazon/s3"
	"github.com/opensds/multi-cloud/datamover/pkg/azure/blob"
	"github.com/opensds/multi-cloud/datamover/pkg/hw/obs"
	"github.com/opensds/multi-cloud/datamover/pkg/ibm/cos"
)

func changeStorageClass(objKey *string, newClass *string, virtBucket *string, bkend *BackendInfo) error {
	if *virtBucket == "" {
		log.Logf("Change storage class of object[%s] is failed: virtual bucket is null.", objKey)
		return errors.New(DMERR_InternalError)
	}

	key := *objKey
	if *virtBucket != "" {
		key = *virtBucket + "/" + *objKey
	}

	var err error
	switch bkend.StorType {
	case flowtype.STOR_TYPE_AWS_S3:
		mover := s3mover.S3Mover{}
		err = mover.ChangeStorageClass(&key, newClass, bkend)
	case flowtype.STOR_TYPE_IBM_COS:
		mover := ibmcosmover.IBMCOSMover{}
		err = mover.ChangeStorageClass(&key, newClass, bkend)
	case flowtype.STOR_TYPE_HW_OBS:
		mover := obsmover.ObsMover{}
		err = mover.ChangeStorageClass(&key, newClass, bkend)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		mover := blobmover.BlobMover{}
		err = mover.ChangeStorageClass(&key, newClass, bkend)
		/*case flowtype.STOR_TYPE_GCP_S3:
			mover := Gcps3mover.GcpS3Mover{}
			err = mover.DeleteObj(objKey, loca)*/
	default:
		log.Logf("Change storage class of object[objkey:%s] failed: backend type is not support.\n", objKey)
		err = errors.New(DMERR_UnSupportBackendType)
	}

	return err
}

func loadStorageClassDefinition() error {
	res, _ := s3client.GetTierMap(context.Background(), &osdss3.NullRequest{})
	if len(res.Tier2Name) == 0 {
		log.Log("Get tier definition failed")
		return fmt.Errorf("Get tier definition failed")
	}

	Int2ExtTierMap = make(map[string]*Int2String)
	for k, v := range res.Tier2Name {
		val := make(Int2String)
		for k1, v1 := range v.Lst {
			val[k1] = v1
		}
		Int2ExtTierMap[k] = &val
	}

	return nil
}

func getStorageClassName(tier int32, storageType string) (string, error) {
	loadFlag := false
	{
		mutex.RLock()
		if len(Int2ExtTierMap) != 0 {
			loadFlag = true
		}
		defer mutex.RUnlock()
	}

	var err error
	if loadFlag == false {
		mutex.Lock()
		defer mutex.Unlock()
		err = loadStorageClassDefinition()
	}
	if err != nil {
		return "", err
	}

	key := ""
	switch storageType {
	case flowtype.STOR_TYPE_AWS_S3:
		key = OSTYPE_AWS
	case flowtype.STOR_TYPE_IBM_COS:
		key = OSTYPE_IBM
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		key = OSTYPE_OBS
	case flowtype.STOR_TYPE_AZURE_BLOB:
		key = OSTYPE_Azure
	case flowtype.STOR_TYPE_CEPH_S3:
		key = OSTYPE_CEPTH
	case flowtype.STOR_TYPE_GCP_S3:
		key = OSTYPE_GCS
	default:
		log.Log("Map tier to storage class name failed: backend type is not support.")
		return "", errors.New(DMERR_UnSupportBackendType)
	}

	className := ""
	v1, _ := Int2ExtTierMap[key]
	v2, ok := (*v1)[tier]
	if !ok {
		err = fmt.Errorf("Tier[%d] is not support for %s", tier, storageType)
	} else {
		className = v2
	}

	return className, err
}

func doInCloudTransition(acReq *datamover.LifecycleActionRequest) error {
	log.Logf("In-cloud transition action: transition %s from %d to %d of %s.\n",
		acReq.ObjKey, acReq.SourceTier, acReq.TargetTier, acReq.TargetBackend)

	loca, err := getBackendInfo(&acReq.BucketName, &acReq.SourceBackend, false)
	if err != nil {
		log.Logf("In-cloud transition of %s failed because get location failed.\n", acReq.ObjKey)
		return err
	}

	className, err := getStorageClassName(acReq.TargetTier, loca.StorType)
	if err != nil {
		log.Logf("In-cloud transition of %s failed because target tier is not supported.\n", acReq.ObjKey)
		return err
	}
	err = changeStorageClass(&acReq.ObjKey, &className, &acReq.BucketName, loca)
	if err != nil && err.Error() == DMERR_NoPermission {
		loca, err = getBackendInfo(&acReq.BucketName, &acReq.SourceBackend, true)
		if err != nil {
			return err
		}
		err = changeStorageClass(&acReq.ObjKey, &className, &acReq.BucketName, loca)
	}

	if err != nil {
		log.Logf("In-cloud transition of %s failed: %v.\n", acReq.ObjKey, err)
		return err
	}

	//update meta data.
	setting := make(map[string]string)
	setting[OBJMETA_TIER] = fmt.Sprintf("%s", acReq.TargetTier)
	//req := osdss3.UpdateObjMetaRequest{acReq.ObjKey, acReq.BucketName}
	req := osdss3.UpdateObjMetaRequest{ObjKey:acReq.ObjKey, BucketName:acReq.BucketName, Setting:setting}
	_, err = s3client.UpdateObjMeta(context.Background(), &req)
	if err != nil {
		log.Logf("Update tier of object[%s] to %d failed:%v.\n", acReq.ObjKey, acReq.TargetTier, err)
		//TODO: Log it and let another module to handle.
	}

	return err
}

