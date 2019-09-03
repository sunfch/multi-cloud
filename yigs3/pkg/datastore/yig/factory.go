package yig

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/opensds/multi-cloud/backend/pkg/utils/constants"
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/driver"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/config"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/redis"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/storage"
)

type YigDriverFactory struct {
	Drivers    sync.Map
	cfgWatcher *config.ConfigWatcher
}

func (ydf *YigDriverFactory) CreateDriver(detail *backendpb.BackendDetail) (driver.StorageDriver, error) {
	// if driver already exists, just return it.
	if driver, ok := ydf.Drivers.Load(detail.Endpoint); ok {
		return driver.(*storage.YigStorage), nil
	}

	return nil, errors.New(fmt.Sprintf("no storage driver for yig endpoint: %s", detail.Endpoint))
}

func (ydf *YigDriverFactory) Init() error {
	// create the driver.
	rand.Seed(time.Now().UnixNano())
	redis.Initialize()

	// init config watcher.
	watcher, err := config.NewConfigWatcher(ydf.driverInit)
	if err != nil {
		return err
	}
	ydf.cfgWatcher = watcher

	// read the config.
	err = config.ReadConfigs("/etc/yig", ydf.driverInit)
	if err != nil {
		return err
	}

	ydf.cfgWatcher.Watch("/etc/yig")
	return nil
}

func (ydf *YigDriverFactory) driverInit(cfg *config.Config) error {
	yigStorage, err := storage.New(cfg)
	if err != nil {
		return err
	}

	ydf.Drivers.Store(cfg.Endpoint.Url, yigStorage)

	return nil
}

func init() {
	yigDf := &YigDriverFactory{}
	err := yigDf.Init()
	if err != nil {
		return
	}
	driver.RegisterDriverFactory(constants.BackendTypeYIGS3, yigDf)
}
