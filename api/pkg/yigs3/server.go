package s3

import (
	"net/http"

	"github.com/opensds/multi-cloud/yigs3/pkg/helper"
	"github.com/opensds/multi-cloud/yigs3/pkg/meta/types"
	"github.com/opensds/multi-cloud/yigs3/pkg/iam/common"
)

const RequestContextKey = "RequestContext"

type RequestContext struct {
	RequestId  string
	BucketInfo *types.Bucket
	ObjectInfo *types.Object
	Credential *common.Credential

}

type Server struct {
	Server *http.Server
}

func (s *Server) Stop() {
	helper.Logger.Print(5, "Stopping API server...")
	helper.Logger.Println(5, "done")
}
