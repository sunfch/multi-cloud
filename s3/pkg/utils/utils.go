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

package utils

type Database struct {
	Credential string `conf:"credential,username:password@tcp(ip:port)/dbname"`
	Driver     string `conf:"driver,mongodb"`
	Endpoint   string `conf:"endpoint,localhost:27017"`
}

const (
	Tier1 = 1
	Tier9 = 99
	Tier99 = 999
)

const (
	AWS_STANDARD = "STANDARD"
	AWS_STANDARD_IA = "STANDARD_IA"
	AWS_GLACIER = "Glacier"
)

const (
	CEPH_STANDARD = "STDANDARD"
)

const (
	GCS_MULTI_REGIONAL = "MULTI_REGIONAL"
	GCS_REGIONAL = "REGIONAL"
	GCS_NEARLINE = "NEARLINE"
	GCS_COLDLINE = "COLDLINE"
)

//Object Storage Type
const (
	OSTYPE_OPENSDS = "OpenSDS"
	OSTYPE_AWS = "AWSS3"
	OSTYPE_Azure = "AzureBlob"
	OSTYPE_OBS = "HuaweiOBS"
	OSTYPE_GCS = "GoogleCloudStorage"
	OSTYPE_CEPTH = "CephS3"
	OSTYPE_FUSIONSTORAGE = "FusionStorage"
)