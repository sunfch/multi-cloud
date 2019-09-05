package tests

import (
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/s3/pkg/datastore/driver"
	. "gopkg.in/check.v1"
)

func (ys *YigSuite) TestPutObjectSucceed(c *C) {
	detail := &backendpb.BackendDetail{
		Endpoint: "default",
	}

	yig, err := driver.CreateStorageDriver("yigs3", detail)
	c.Assert(err, Equals, nil)
	c.Assert(yig, Not(Equals), nil)
}
