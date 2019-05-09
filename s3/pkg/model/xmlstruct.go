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

package model

var Xmlns = "http://s3.amazonaws.com/doc/2006-03-01"

type CreateBucketConfiguration struct {
	Xmlns              string `xml:"xmlns,attr"`
	LocationConstraint string `xml:"LocationConstraint"`
}

type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type Bucket struct {
	Name               string `xml:"Name"`
	CreationDate       string `xml:"CreationDate"`
	LocationConstraint string `xml:"LocationConstraint"`
}

type ListAllMyBucketsResult struct {
	Xmlns   string   `xml:"xmlns,attr"`
	Owner   Owner    `xml:"Owner"`
	Buckets []Bucket `xml:"Buckets"`
}

type InitiateMultipartUploadResult struct {
	Xmlns    string `xml:"xmlns,attr"`
	Bucket   string `xml:"Bucket"`
	Key      string `xml:"Key"`
	UploadId string `xml:"UploadId"`
}

type UploadPartResult struct {
	Xmlns      string `xml:"xmlns,attr"`
	PartNumber int64  `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

type Part struct {
	PartNumber int64  `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

type CompleteMultipartUpload struct {
	Xmlns string `xml:"xmlns,attr"`
	Part  []Part `xml:"Part"`
}

type CompleteMultipartUploadResult struct {
	Xmlns    string `xml:"xmlns,attr"`
	Location string `xml:"Location"`
	Bucket   string `xml:"Bucket"`
	Key      string `xml:"Key"`
	ETag     string `xml:"ETag"`
}

type ListPartsOutput struct {
	Xmlns       string `xml:"xmlns,attr"`
	Bucket      string `xml:"Bucket"`
	Key         string `xml:"Key"`
	UploadId    string `xml:"UploadId"`
	MaxParts    int    `xml:"MaxParts"`
	IsTruncated bool   `xml:"IsTruncated"`
	Owner       Owner  `xml:"Owner"`
	Parts       []Part `xml:"Part"`
}

type StorageClass struct {
	Name               string `xml:"Name"`
	Tier               int32 `xml:"Tier"`
}

type ListStorageClasses struct {
	Xmlns       string `xml:"xmlns,attr"`
	Classes     []StorageClass `xml:"Class"`
}
