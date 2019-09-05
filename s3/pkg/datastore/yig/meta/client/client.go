package client

import (
	"github.com/opensds/multi-cloud/s3/pkg/datastore/yig/meta/types"
)

//DB Client Interface
type Client interface {
	//Transaction
	NewTrans() (tx interface{}, err error)
	AbortTrans(tx interface{}) error
	CommitTrans(tx interface{}) error

	//cluster
	GetCluster(fsid, pool string) (cluster types.Cluster, err error)
}
