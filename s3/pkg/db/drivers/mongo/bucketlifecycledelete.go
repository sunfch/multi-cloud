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

package mongo

import (
	"context"

	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-log"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	. "github.com/opensds/multi-cloud/s3/pkg/utils"
	pb "github.com/opensds/multi-cloud/s3/proto"
)

func (ad *adapter) DeleteBucketLifecycle(ctx context.Context, in *pb.DeleteLifecycleInput) S3Error {
	//Check if the connector exist or not
	ss := ad.s.Copy()
	defer ss.Close()

	log.Logf("delete bucket lifecycle, bucketName is %v, lifecycle id is %s\n", in.Bucket, in.RuleID)

	m := bson.M{DBKEY_NAME: in.Bucket}
	err := UpdateContextFilter(ctx, m)
	if err != nil {
		return InternalError
	}

	//Delete it from database
	err = ss.DB(DataBaseName).C(BucketMD).Update(m, bson.M{"$pull": bson.M{DBKEY_LIFECYCLE:
		bson.M{DBKEY_ID: in.RuleID}}})
	if err != nil {
		log.Logf("delete lifecycle for bucket : %s and lifecycle ruleID : %s failed,err:%v.\n",
			in.Bucket, in.RuleID, err)
		return NoSuchBucket
	} else {
		log.Logf("delete bucket lifecycle with rule id %s from database successfully", in.RuleID)
		return NoError
	}
}
