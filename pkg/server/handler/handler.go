/*
Copyright 2023 EscherCloudAI.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler

import (
	"github.com/eschercloudai/baski/pkg/s3"
	"github.com/eschercloudai/baski/pkg/server/generated"
	"github.com/eschercloudai/baski/pkg/server/handler/scans"
	"github.com/eschercloudai/baski/pkg/server/server/util"
	"net/http"
)

type Handler struct {
	s3 *s3.S3
}

func New(endpoint, access, secret, bucket string) *Handler {
	h := &Handler{
		s3: &s3.S3{
			Endpoint:  endpoint,
			AccessKey: access,
			SecretKey: secret,
			Bucket:    bucket,
		},
	}

	return h
}

func (h *Handler) ApiV1GetScan(w http.ResponseWriter, r *http.Request, imageId generated.ImageID) {
	res, err := scans.FetchScanResult(imageId, h.s3)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}
	util.JSON(w, http.StatusOK, res)
}
