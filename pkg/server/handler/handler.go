/*
Copyright 2023 EscherCloud.

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
	"encoding/json"
	"fmt"
	"github.com/eschercloudai/baski/pkg/server/generated"
	"github.com/eschercloudai/baski/pkg/server/server/util"
	u "github.com/eschercloudai/baski/pkg/util"
	"net/http"
)

type Handler struct {
	s3 u.S3Interface
}

func New(s u.S3Interface) *Handler {
	h := &Handler{
		s3: s,
	}

	return h
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		util.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
}

func (h *Handler) ApiV1GetScan(w http.ResponseWriter, r *http.Request, imageId generated.ImageID) {
	res, err := h.s3.FetchFromS3(fmt.Sprintf("scans/%s/results.json", imageId))
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}
	scanRes := []generated.ScanResult{}

	err = json.Unmarshal(res, &scanRes)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}
	util.JSON(w, http.StatusOK, scanRes)
}
