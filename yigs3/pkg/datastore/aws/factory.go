package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opensds/multi-cloud/backend/pkg/utils/constants"
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/driver"
)

type AwsS3DriverFactory struct {
}

func (as3 *AwsS3DriverFactory) CreateDriver(detail *backendpb.BackendDetail) (driver.StorageDriver, error) {
	endpoint := detail.Endpoint
	AccessKeyID := detail.Access
	AccessKeySecret := detail.Security
	region := detail.Region

	s3aksk := s3Cred{ak: AccessKeyID, sk: AccessKeySecret}
	creds := credentials.NewCredentials(&s3aksk)

	disableSSL := true
	sess, err := session.NewSession(&aws.Config{
		Region:      &region,
		Endpoint:    &endpoint,
		Credentials: creds,
		DisableSSL:  &disableSSL,
	})
	if err != nil {
		return nil, err
	}

	adap := &AwsAdapter{backend: detail, session: sess}

	return adap, nil
}

func init() {
	driver.RegisterDriverFactory(constants.BackendTypeAws, &AwsS3DriverFactory{})
}
