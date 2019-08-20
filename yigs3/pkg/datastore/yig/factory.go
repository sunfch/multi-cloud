package yig

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	backendpb "github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/driver"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/helper"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/log"
	bus "github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/messagebus"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/redis"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/storage"
)

const (
	YIG_DRIVER_TYPE = "yig-s3"
)

type YigDriverFactory struct {
	Drivers sync.Map
}

func (ydf *YigDriverFactory) CreateDriver(detail *backendpb.BackendDetail) (driver.StorageDriver, error) {
	// if driver already exists, just return it.
	if driver, ok := ydf.Drivers.Load("default" /*detail.Endpoint*/); ok {
		return driver.(*storage.YigStorage), nil
	}

	return nil, errors.New(fmt.Sprintf("no storage driver for yig endpoint: %s", detail.Endpoint))
}

func (ydf *YigDriverFactory) init() {
	// create the driver.
	rand.Seed(time.Now().UnixNano())

	helper.SetupConfig()

	//yig log
	f, err := os.OpenFile(helper.CONFIG.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Failed to open log file " + helper.CONFIG.LogPath)
	}
	defer f.Close()

	logger := log.New(f, "[yig]", log.LstdFlags, helper.CONFIG.LogLevel)
	helper.Logger = logger
	logger.Printf(20, "YIG conf: %+v \n", helper.CONFIG)
	logger.Println(5, "YIG instance ID:", helper.CONFIG.InstanceId)

	if helper.CONFIG.MetaCacheType > 0 || helper.CONFIG.EnableDataCache {
		redis.Initialize()
	}

	yigStorage := storage.New(logger, helper.CONFIG.MetaCacheType, helper.CONFIG.EnableDataCache, helper.CONFIG.CephConfigPattern)

	// try to create message bus sender if message bus is enabled.
	// message bus sender is singleton so create it beforehand.
	if helper.CONFIG.MsgBus.Enabled {
		messageBusSender, err := bus.GetMessageSender()
		if err != nil {
			helper.Logger.Printf(2, "failed to create message bus sender, err: %v", err)
			panic("failed to create message bus sender")
		}
		if nil == messageBusSender {
			helper.Logger.Printf(2, "failed to create message bus sender, sender is nil.")
			panic("failed to create message bus sender, sender is nil.")
		}
		helper.Logger.Printf(20, "succeed to create message bus sender.")
	}

	ydf.Drivers.Store("default", yigStorage)
}

func init() {
	driver.RegisterDriverFactory(YIG_DRIVER_TYPE, &YigDriverFactory{})
}
