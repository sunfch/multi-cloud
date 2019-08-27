package s3

import (
	"context"
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	"github.com/micro/go-micro/metadata"
	"github.com/opensds/multi-cloud/api/pkg/common"
	c "github.com/opensds/multi-cloud/api/pkg/context"
	"github.com/opensds/multi-cloud/api/pkg/s3/datastore"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	s3 "github.com/opensds/multi-cloud/s3/proto"
)

func (s *APIService) AbortMultipartUpload(request *restful.Request, response *restful.Response) {
	bucketName := request.PathParameter("bucketName")
	objectKey := request.PathParameter("objectKey")
	uploadId := request.QueryParameter("uploadId")

	actx := request.Attribute(c.KContext).(*c.Context)
	ctx := metadata.NewContext(context.Background(), map[string]string{
		common.CTX_KEY_USER_ID:    actx.UserId,
		common.CTX_KEY_TENENT_ID:  actx.TenantId,
		common.CTX_KEY_IS_ADMIN:   strconv.FormatBool(actx.IsAdmin),
		common.REST_KEY_OPERATION: common.REST_VAL_MULTIPARTUPLOAD,
	})

	objectInput := s3.GetObjectInput{Bucket: bucketName, Key: objectKey}
	objectMD, _ := s.s3Client.GetObject(ctx, &objectInput)
	multipartUpload := s3.MultipartUpload{}
	multipartUpload.Key = objectKey
	multipartUpload.Bucket = bucketName
	multipartUpload.UploadId = uploadId

	var client datastore.DataStoreAdapter
	if objectMD == nil {
		log.Logf("no such object err\n")
		response.WriteError(http.StatusInternalServerError, NoSuchObject.Error())

	}
	client = getBackendByName(ctx, s, objectMD.Backend)
	if client == nil {
		response.WriteError(http.StatusInternalServerError, NoSuchBackend.Error())
		return
	}
	s3err := client.AbortMultipartUpload(&multipartUpload, ctx)
	if s3err != NoError {
		response.WriteError(http.StatusInternalServerError, s3err.Error())
		return
	}

	// delete multipart upload record, if delete failed, it will be cleaned by lifecycle management
	record := s3.MultipartUploadRecord{ObjectKey: objectKey, Bucket: bucketName, UploadId: uploadId}
	s.s3Client.DeleteUploadRecord(context.Background(), &record)

	deleteInput := s3.DeleteObjectInput{Key: objectKey, Bucket: bucketName}
	res, err := s.s3Client.DeleteObject(ctx, &deleteInput)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	log.Logf("Delete object %s successfully.", objectKey)
	response.WriteEntity(res)
}
