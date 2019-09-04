package meta

import (
	"errors"

	"github.com/opensds/multi-cloud/yigs3/pkg/helper"
	"github.com/opensds/multi-cloud/yigs3/pkg/log"
	"github.com/opensds/multi-cloud/yigs3/pkg/meta/client"
	"github.com/opensds/multi-cloud/yigs3/pkg/meta/client/tidbclient"
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

func (m *Meta) Sync(event SyncEvent) error {
	switch event.Type {
	case SYNC_EVENT_TYPE_BUCKET_USAGE:
		return m.bucketUsageSync(event)
	default:
		return errors.New("got unknown sync event.")
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
	err := meta.InitBucketUsageCache()
	if err != nil {
		panic("failed to init bucket usage cache")
	}
	return &meta
}
