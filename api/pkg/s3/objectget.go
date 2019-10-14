// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3

import (
	"io"

	"github.com/emicklei/go-restful"
	"github.com/opensds/multi-cloud/api/pkg/common"
	log "github.com/sirupsen/logrus"
	pb "github.com/opensds/multi-cloud/s3/proto"
	/*c "github.com/opensds/multi-cloud/api/pkg/context"
	. "github.com/opensds/multi-cloud/s3/pkg/exception"
	s3 "github.com/opensds/multi-cloud/s3/proto"
	"golang.org/x/net/context"
	*/
)

// Simple way to convert a func to io.Writer type.
type funcToWriter func([]byte) (int, error)

func (f funcToWriter) Write(p []byte) (int, error) {
	return f(p)
}

//ObjectGet -
func (s *APIService) ObjectGet(request *restful.Request, response *restful.Response) {
	bucketName := request.PathParameter("bucketName")
	objectKey := request.PathParameter("objectKey")
	rangestr := request.HeaderParameter("Range")
	log.Infof("%v\n", rangestr)

	ctx := common.InitCtxWithAuthInfo(request)
	input := &pb.GetObjectInput{
		Bucket:bucketName,
		Key:objectKey,
		Versionid:"",
	}

	object, err := s.s3Client.GetObjectInfo(ctx, input)
	if err != nil {
		log.Errorln("failed to call s3 GetObjectInfo")
		return
	}
	if object.DeleteMarker {
		log.Infoln("the object is delete marker.")
		return
	}
	// Indicates if any data was written to the http.ResponseWriter
	dataWritten := false

	// io.Writer type which keeps track if any data was written.
	writer := funcToWriter(func(p []byte) (int, error) {
		if !dataWritten {
			// Set headers on the first write.
			// Set standard object headers.
			//SetObjectHeaders(w, object, hrange)

			// Set any additional requested response headers.
			//setGetRespHeaders(w, r.URL.Query())

			/*if version != "" {
				w.Header().Set("x-amz-version-id", version)
			}*/
			dataWritten = true
		}
		n, err := response.Write(p)
		if n > 0 {
			/*
				If the whole write or only part of write is successfull,
				n should be positive, so record this
			*/
			//w.(*ResponseRecorder).size += int64(n)
		}
		return n, err
	})

	stream, err := s.s3Client.GetObject(ctx, object)
	if err != nil {
		log.Errorln("failed to call s3 GetObject")
		return
	}
	defer stream.Close()

	eof := false
	for !eof {
		rsp, err := stream.Recv()
		if err != nil && err != io.EOF {
			log.Errorln("recv err", err)
			break
		}
		if err == io.EOF {
			eof = true
		}
		_, err = writer.Write(rsp.Data)
		if err != nil {
			log.Errorln("failed to write data to client. err:", err)
			break
		}
	}

	if !dataWritten {
		// If ObjectAPI.GetObject did not return error and no data has
		// been written it would mean that it is a 0-byte object.
		// call wrter.Write(nil) to set appropriate headers.
		writer.Write(nil)
	}

	log.Info("GET object successfully.")
}
