package utils

type Database struct {
	Credential string `conf:"credential,username:password@tcp(ip:port)/dbname"`
	Driver     string `conf:"driver,mongodb"`
	Endpoint   string `conf:"endpoint,localhost:27017"`
}

const (
	Tier1 = 1
	Tier9 = 99
	Tier99 = 999
)

const (
	AWS_STANDARD = "STANDARD"
	AWS_STANDARD_IA = "STANDARD_IA"
	AWS_GLACIER = "Glacier"
)

const (
	CEPH_STANDARD = "STDANDARD"
)

const (
	GCS_MULTI_REGIONAL = "MULTI_REGIONAL"
	GCS_REGIONAL = "REGIONAL"
	GCS_NEARLINE = "NEARLINE"
	GCS_COLDLINE = "COLDLINE"
)

//Object Storage Type
const (
	OSTYPE_AWS = "AWSS3"
	OSTYPE_Azure = "AzureBlob"
	OSTYPE_OBS = "HuaweiOBS"
	OSTYPE_GCS = "GoogleCloudStorage"
	OSTYPE_CEPTH = "CephS3"
	OSTYPE_FUSIONSTORAGE = "FusionStorage"
)