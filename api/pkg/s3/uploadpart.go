package s3

import (
	"context"
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	"github.com/opensds/multi-cloud/api/pkg/s3/datastore"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-log"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	s3 "github.com/opensds/multi-cloud/s3/proto"
)

func (s *APIService) UploadPart(request *restful.Request, response *restful.Response) {
	bucketName := request.PathParameter("bucketName")
	objectKey := request.PathParameter("objectKey")
	contentLenght := request.HeaderParameter("content-length")
	size, _ := strconv.ParseInt(contentLenght, 10, 64)

	uploadId := request.QueryParameter("uploadId")
	partNumber := request.QueryParameter("partNumber")
	partNumberInt, _ := strconv.ParseInt(partNumber, 10, 64)
	ctx := context.WithValue(request.Request.Context(), "operation", "multipartupload")
	objectInput := s3.GetObjectInput{Bucket: bucketName, Key: objectKey}
	objectMD, _ := s.s3Client.GetObject(ctx, &objectInput)
	lastModified := time.Now().Unix()
	object := s3.Object{}
	object.ObjectKey = objectKey
	object.BucketName = bucketName
	object.LastModified = lastModified
	object.Size = size
	var client datastore.DataStoreAdapter
	if objectMD == nil {
		log.Logf("No such object err\n")
		response.WriteError(http.StatusInternalServerError, NoSuchObject.Error())

	}
	log.Logf("objectMD.Backend is %v\n", objectMD.Backend)
	client = getBackendByName(s, objectMD.Backend)
	if client == nil {
		response.WriteError(http.StatusInternalServerError, NoSuchBackend.Error())
		return
	}
	multipartUpload := s3.MultipartUpload{}
	multipartUpload.Bucket = bucketName
	multipartUpload.Key = objectKey
	multipartUpload.UploadId = uploadId
	log.Logf("call .UploadPart api")
	//call API
	res, s3err := client.UploadPart(request.Request.Body, &multipartUpload, partNumberInt, request.Request.ContentLength, ctx)
	if s3err != NoError {
		response.WriteError(http.StatusInternalServerError, s3err.Error())
		return
	}

	partion := s3.Partion{}

	partion.PartNumber = partNumber
	log.Logf("uploadPart size is %v", size)
	partion.Size = size
	timestamp := time.Now().Unix()
	partion.LastModified = timestamp
	partion.Key = objectKey
	log.Logf("objectMD.Size1 = %v", objectMD.Size)
	objectMD.Size = objectMD.Size + size
	log.Logf("objectMD.Size2 = %v", objectMD.Size)
	objectMD.LastModified = lastModified
	objectMD.Partions = append(objectMD.Partions, &partion)
	//insert metadata
	_, err := s.s3Client.CreateObject(ctx, objectMD)
	result, _ := s.s3Client.GetObject(ctx, &objectInput)
	log.Logf("result.size = %v", result.Size)
	if err != nil {
		log.Logf("err is %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
	}

	//return xml format
	xmlstring, err := xml.MarshalIndent(res, "", "  ")
	if err != nil {
		log.Logf("Parse ListBuckets error: %v", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	xmlstring = []byte(xml.Header + string(xmlstring))
	log.Logf("resp:\n%s", xmlstring)
	response.Write(xmlstring)

	log.Log("Uploadpart successfully.")
}
