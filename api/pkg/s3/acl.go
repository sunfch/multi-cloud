package s3

import (
	. "github.com/opensds/multi-cloud/api/pkg/s3/datatype"
	"net/http"
)

func getAclFromHeader(h http.Header) (acl Acl, err error) {
	acl.CannedAcl = h.Get("x-amz-acl")
	if acl.CannedAcl == "" {
		acl.CannedAcl = "private"
	}
	err = IsValidCannedAcl(acl)
	return
}
