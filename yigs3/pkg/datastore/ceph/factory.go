package ceph

import (
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore"
	"github.com/webrtcn/s3client"
)

const (
	CEPH_DRIVER_TYPE = "ceph-s3"
)

type CephDriverFactory struct {
}

func (cdf *CephDriverFactory) CreateDriver(detail *backendpb.BackendDetail) (datastore.StorageDriver, error) {
	endpoint := detail.Endpoint
	AccessKeyID := detail.Access
	AccessKeySecret := detail.Security
	sess := s3client.NewClient(endpoint, AccessKeyID, AccessKeySecret)
	adap := &CephAdapter{backend: detail, session: sess}
	return adap, nil
}

func init() {
	datastore.RegisterDriverFactory(CEPH_DRIVER_TYPE, &CephDriverFactory{})
}
