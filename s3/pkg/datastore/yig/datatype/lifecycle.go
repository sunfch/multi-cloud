package datatype

import (
	"encoding/xml"
	//	. "github.com/opensds/multi-cloud/s3/pkg/datastore/yig/error"
	//	"github.com/opensds/multi-cloud/s3/pkg/helper"
)

type LcRule struct {
	ID         string `xml:"ID"`
	Prefix     string `xml:"Prefix"`
	Status     string `xml:"Status"`
	Expiration string `xml:"Expiration>Days"`
}

type Lc struct {
	XMLName xml.Name `xml:"LifecycleConfiguration"`
	Rule    []LcRule `xml:"Rule"`
}
