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

package cephs3mover

import (
	"errors"
	"github.com/go-log/log"
	. "github.com/opensds/multi-cloud/datamover/pkg/utils"
)

func (mover *CephS3Mover)ChangeStorageClass(objKey *string, newClass *string, bkend *BackendInfo) error {
	log.Logf("[cephs3lifecycle]: Failed to change storage class of object[key:%s, backend:%s] to %s, no transition support for ceph s3.\n",
		objKey, bkend.BakendName, newClass)

	return errors.New(DMERR_UnSupportOperation)
}
