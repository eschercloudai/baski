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
	"fmt"
	"github.com/eschercloudai/baski/pkg/mock"
	"github.com/eschercloudai/baski/pkg/s3"
	th "github.com/eschercloudai/baski/testhelpers"
	"go.uber.org/mock/gomock"
	"io"
	"log"
	"net/http"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	endpoint := "test"
	cloud := "testCloud"
	access := "abc"
	secret := "def"
	bucket := "a_bucket"

	b, err := s3.New(endpoint, access, secret, bucket, "")
	if err != nil {
		log.Println(err)
		return
	}
	d, err := s3.New(endpoint, access, secret, bucket, "")
	if err != nil {
		log.Println(err)
		return
	}
	expected := &Handler{
		baskiS3:   b,
		dogkatS3:  d,
		cloudName: cloud,
	}

	test := New(b, d, cloud)
	if !reflect.DeepEqual(expected, test) {
		t.Errorf("expected %+v, got %+v", expected, test)
	}
}

func TestApiV1GetScan(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	expectedResponse := []byte(`[{"cvss":{"ghsa":{"v2Score":"string","v2Vector":"string","v3Score":"string","v3Vector":"string"},"nvd":{"v2Score":"string","v2Vector":"string","v3Score":"string","v3Vector":"string"},"redhat":{"v2Score":"string","v2Vector":"string","v3Score":"string","v3Vector":"string"}},"description":"string","fixedVersion":"string","installedVersion":"string","pkgName":"string","severity":"string","vulnerabilityID":"string"}]`)

	c := gomock.NewController(t)
	defer c.Finish()
	bm := mock.NewMockS3Interface(c)

	bm.EXPECT().Fetch(gomock.Eq("scans/abcde-123456/results.json")).Return(expectedResponse, nil)

	h := Handler{baskiS3: bm}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		h.ApiV1GetScan(w, r, "abcde-123456")
	}

	th.Mux.HandleFunc("/api/v1/scan/abcde-123456", handler)

	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/scan/%s", th.Port, "abcde-123456"))
	if err != nil {
		t.Error(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if string(expectedResponse) != string(body) {
		t.Errorf("expected %s, got %s", expectedResponse, body)
	}
}

func TestApiV1GetTest(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	expectedResponse := []byte(`[{"completed":"true","description":"Some description","name":"some_test"},{"completed":"true","description":"Some description2","name":"some_test2"}]`)

	c := gomock.NewController(t)
	defer c.Finish()
	dm := mock.NewMockS3Interface(c)

	dm.EXPECT().Fetch(gomock.Eq("abcde-123456.json")).Return(expectedResponse, nil)

	h := Handler{dogkatS3: dm}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		h.ApiV1GetTest(w, r, "abcde-123456")
	}

	th.Mux.HandleFunc("/api/v1/test/abcde-123456", handler)

	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/test/%s", th.Port, "abcde-123456"))
	if err != nil {
		t.Error(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if string(expectedResponse) != string(body) {
		t.Errorf("expected %s, got %s", expectedResponse, body)
	}
}

func TestHealthz(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	expectedResponse := []byte(`ok`)

	h := Handler{}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		h.Healthz(w, r)
	}

	th.Mux.HandleFunc("/healthz", handler)

	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", th.Port))
	if err != nil {
		t.Error(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if string(expectedResponse) != string(body) {
		t.Errorf("expected %s, got %s", expectedResponse, body)
	}
}
