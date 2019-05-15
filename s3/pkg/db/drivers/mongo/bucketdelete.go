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

package mongo

import (
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-log"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
)

func (ad *adapter) DeleteBucket(bucketName string) S3Error {
	//Check if the connctor exist or not
	ss := ad.s.Copy()
	defer ss.Close()

	//Delete it from database
	c := ss.DB(DataBaseName).C(BucketMD)
	log.Logf("bucketName is %v:", bucketName)
	err := c.Remove(bson.M{"name": bucketName})
	log.Logf("err is %v:", err)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Logf("Delete bucket from database failed,err:%v.\n", err.Error())
			return NoSuchBucket
		} else {
			log.Logf("Delete bucket from database failed,err:%v.\n", err.Error())
			return DBError
		}

	} else {
		log.Logf("Delete bucket from database successfully")
		return NoError
	}
	cc := ss.DB(DataBaseName).C(bucketName)
	deleteErr := cc.DropCollection()
	if deleteErr != nil && deleteErr != mgo.ErrNotFound {
		log.Logf("Delete bucket collection from database failed,err:%v.\n", deleteErr)
		return InternalError
	}

	return NoError
}

func (ad *adapter) DeleteLifecycle(bucketName string, lifecycleId string) S3Error {
	ss := ad.s.Copy()
	defer ss.Close()

	//Update database
	c := ss.DB(DataBaseName).C(BucketMD)
	err := c.Update(bson.M{"name": bucketName}, bson.M{"$pull": bson.M{"lifecycleconfiguration": bson.M{"id": lifecycleId}}})
	if err == mgo.ErrNotFound {
		log.Logf("delete lifecycle[id=%s] of bucket[%s] failed, err:the specified bucket or lifecycle rule does not exist.",
			lifecycleId, bucketName)
		return NoSuchLifecycleRule
	} else if err != nil {
		log.Logf("delete lifecycle[bucket=%s, lifecycleid=%s] failed: %v\n", bucketName, lifecycleId, err)
		return DBError
	}

	log.Logf("delete lifecycle[bucket=%s, lifecycleid=%s] successfully\n", bucketName, lifecycleId)
	return NoError
}