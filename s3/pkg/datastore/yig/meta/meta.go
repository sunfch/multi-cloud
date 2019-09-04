package meta

import (
	"github.com/opensds/multi-cloud/s3/pkg/datastore/yig/helper"
	"github.com/opensds/multi-cloud/s3/pkg/datastore/yig/log"
	"github.com/opensds/multi-cloud/s3/pkg/datastore/yig/meta/client"
	"github.com/opensds/multi-cloud/s3/pkg/datastore/yig/meta/client/tidbclient"
)

const (
	ENCRYPTION_KEY_LENGTH = 32 // 32 bytes for AES-"256"
)

type Meta struct {
	Client client.Client
	Logger *log.Logger
	Cache  MetaCache
}

func (m *Meta) Stop() {
	if m.Cache != nil {
		m.Cache.Close()
	}
}

func New(logger *log.Logger, myCacheType CacheType) *Meta {
	meta := Meta{
		Logger: logger,
		Cache:  newMetaCache(myCacheType),
	}
	if helper.CONFIG.MetaStore == "tidb" {
		meta.Client = tidbclient.NewTidbClient()
	} else {
		panic("unsupport metastore")
	}
	return &meta
}
