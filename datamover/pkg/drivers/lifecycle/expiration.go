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
	"context"
	"github.com/micro/go-log"
	osdss3 "github.com/opensds/multi-cloud/s3/proto"
	flowtype "github.com/opensds/multi-cloud/dataflow/pkg/model"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	"github.com/opensds/multi-cloud/datamover/proto"
	"github.com/opensds/multi-cloud/datamover/pkg/amazon/s3"
	"github.com/opensds/multi-cloud/datamover/pkg/azure/blob"
	"github.com/opensds/multi-cloud/datamover/pkg/ceph/s3"
	"github.com/opensds/multi-cloud/datamover/pkg/hw/obs"
	"github.com/opensds/multi-cloud/datamover/pkg/gcp/s3"
	"github.com/opensds/multi-cloud/datamover/pkg/ibm/cos"
)

func deleteObj(objKey string, virtBucket string, bkend *BackendInfo) error {
	if virtBucket == "" {
		log.Logf("Expiration of object[%s] is failed: virtual bucket is null.", objKey)
		return errors.New(DMERR_InternalError)
	}

	//delete metadata
	delMetaReq := osdss3.DeleteObjectInput{Bucket: virtBucket, Key: objKey}
	ctx := context.Background()
	_, err := s3client.DeleteObject(ctx, &delMetaReq)
	if err != nil {
		log.Logf("Delete object metadata of obj[bucket:%s,objKey:%s] failed, err:%v\n",
			virtBucket, objKey, err)
		return err
	} else {
		log.Logf("Delete object metadata of obj[bucket:%s,objKey:%s] successfully.\n",
			virtBucket,	objKey)
	}

	if virtBucket != "" {
		objKey = virtBucket + "/" + objKey
	}

	loca := &LocationInfo{StorType:bkend.StorType, Region:bkend.Region, EndPoint:bkend.EndPoint, BucketName: virtBucket,
		Access:bkend.Access, Security:bkend.Security, BakendName:bkend.BakendName}
	switch bkend.StorType {
	case flowtype.STOR_TYPE_AWS_S3:
		mover := s3mover.S3Mover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_IBM_COS:
		mover := ibmcosmover.IBMCOSMover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_HW_OBS, flowtype.STOR_TYPE_HW_FUSIONSTORAGE, flowtype.STOR_TYPE_HW_FUSIONCLOUD:
		mover := obsmover.ObsMover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_AZURE_BLOB:
		mover := blobmover.BlobMover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_CEPH_S3:
		mover := cephs3mover.CephS3Mover{}
		err = mover.DeleteObj(objKey, loca)
	case flowtype.STOR_TYPE_GCP_S3:
		mover := Gcps3mover.GcpS3Mover{}
		err = mover.DeleteObj(objKey, loca)
	default:
		log.Logf("Delete object[objkey:%s] from backend storage failed: backend type is not support.\n", objKey)
		err = errors.New(DMERR_UnSupportBackendType)
	}

	if err != nil {
		log.Logf("Delete object[objkey:%s] from backend storage failed: %v\n", objKey, err)
		//TODO: Log it and let another module to handle.
	}

	return err
}

func doExpirationAction(acReq *datamover.LifecycleActionRequest) error {
	log.Logf("Delete action: delete %s.\n", acReq.ObjKey)

	loc, err := getBackendInfo(&acReq.BucketName, &acReq.SourceBackend, false)
	if err != nil {
		log.Logf("Exipration of %s failed because get location failed.\n", acReq.ObjKey)
	}

	err = deleteObj(acReq.ObjKey, acReq.BucketName, loc)
	if err != nil && err.Error() == DMERR_NoPermission {
		loc, err = getBackendInfo(&acReq.BucketName, &acReq.SourceBackend, true)
		if err != nil {
			return err
		}
		err = deleteObj(acReq.ObjKey, acReq.BucketName, loc)
	}

	return err
}

