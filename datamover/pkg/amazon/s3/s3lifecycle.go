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

package s3mover

import (
	"github.com/micro/go-log"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func (mover *S3Mover)ChangeStorageClass(objKey *string, newClass *string, loca *BackendInfo) error {
	log.Logf("[s3lifecycle] Change storage class of object[%s] to %s.", objKey, newClass)
	s3c := s3Cred{ak: loca.Access, sk: loca.Security}
	creds := credentials.NewCredentials(&s3c)
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(loca.Region),
		Endpoint:    aws.String(loca.EndPoint),
		Credentials: creds,
	})
	if err != nil {
		log.Logf("[s3lifecycle] new session failed, err:%v\n", err)
		return handleAWSS3Errors(err)
	}

	svc := s3.New(sess)
	input := &s3.CopyObjectInput{
		Bucket: aws.String(loca.BucketName),
		Key: aws.String(*objKey),
		CopySource: aws.String(loca.BucketName + "/" + *objKey),
	}
	input.StorageClass = aws.String(*newClass)
	_, err = svc.CopyObject(input)
	if err != nil {
		log.Logf("[s3lifecycle] Change storage class of object[%s] to %s failed: %v.\n", objKey, newClass, err)
		e := handleAWSS3Errors(err)
		return e
	}

	// TODO: How to make sure copy is complemented? Wait to see if the item got copied (example:svc.WaitUntilObjectExists)?

	return nil
}
