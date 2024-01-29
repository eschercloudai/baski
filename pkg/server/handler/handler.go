/*
Copyright 2024 Drewbernetes.

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
	ostack "github.com/drewbernetes/baski/pkg/providers/openstack"
	"github.com/drewbernetes/baski/pkg/server/generated"
	"github.com/drewbernetes/baski/pkg/server/server/util"
	u "github.com/drewbernetes/baski/pkg/util"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
	baskiS3   u.S3Interface
	dogkatS3  u.S3Interface
	cloudName string
}

type ImageData struct {
	ImageID    string               `json:"image_id"`
	ImageName  string               `json:"image_name"`
	K8SVersion interface{}          `json:"k8s_version"`
	BuildDate  interface{}          `json:"build_date"`
	ScanStatus interface{}          `json:"scan_status"`
	ScanData   generated.ScanResult `json:"scan_data"`
	TestStatus interface{}          `json:"test_status"`
	TestData   generated.TestResult `json:"test_data"`
}

func New(b, d u.S3Interface, cloudName string) *Handler {
	h := &Handler{
		baskiS3:   b,
		dogkatS3:  d,
		cloudName: cloudName,
	}

	return h
}

// Healthz returns ok if everything is ok
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		util.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
}

// ApiV1GetScans returns a list of scan results by listing the contents of the S3 bucket.
// This allows for parsing and pulling subsequent data based off of the image ID
func (h *Handler) ApiV1GetScans(w http.ResponseWriter, r *http.Request) {
	data, err := fetchScansList(h)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	util.JSON(w, http.StatusOK, data)
}

// fetchScanData grabs the scan data from S3 and parses it into JSON
func fetchScansList(h *Handler) ([]string, error) {
	contents, err := h.baskiS3.List()
	if err != nil {
		return nil, err
	}

	data := []string{}

	for _, v := range contents {
		split := strings.Split(v, "/")
		if split[0] == "scans" {
			if split[1] != "" {
				data = append(data, split[1])
			}
		}
	}

	return data, nil

}

// ApiV1GetScan will return the raw scan data straight from S3
func (h *Handler) ApiV1GetScan(w http.ResponseWriter, r *http.Request, imageId generated.ImageID) {
	res, err := fetchScanData(h, imageId)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	util.JSON(w, http.StatusOK, res)
}

// fetchScanData grabs the scan data from S3 and parses it into JSON
func fetchScanData(h *Handler, imageId generated.ImageID) (generated.ScanResult, error) {
	res, err := h.baskiS3.Fetch(fmt.Sprintf("scans/%s/results.json", imageId))
	if err != nil {
		return nil, err
	}

	scanRes := generated.ScanResult{}

	if string(res) == "{}" {
		return scanRes, nil
	}
	err = json.Unmarshal(res, &scanRes)
	if err != nil {
		return nil, err
	}

	return scanRes, nil
}

// ApiV1GetTest will return the raw test data straight from S3
func (h *Handler) ApiV1GetTest(w http.ResponseWriter, r *http.Request, imageId generated.ImageID) {
	if h.dogkatS3 == nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": "dogkat support is not enabled on the server"})
		return
	}
	res, err := fetchTestData(h, imageId)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}
	util.JSON(w, http.StatusOK, res)
}

// fetchTestData grabs the test data from S3 and parses it into JSON
func fetchTestData(h *Handler, imageId generated.ImageID) (generated.TestResult, error) {
	res, err := h.dogkatS3.Fetch(fmt.Sprintf("%s.json", imageId))
	if err != nil {
		return nil, err
	}

	testRes := generated.TestResult{}

	if string(res) == "{}" {
		return testRes, nil
	}
	err = json.Unmarshal(res, &testRes)
	if err != nil {
		return nil, err
	}

	return testRes, nil
}

// ApiV1GetImages returns a list of scan results by listing the contents of the S3 bucket.
// This then gets used to collect together image data per image from OpenStack
func (h *Handler) ApiV1GetImages(w http.ResponseWriter, r *http.Request) {
	contents, err := fetchScansList(h)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	cloudProvider := ostack.NewCloudsProvider(h.cloudName)

	i, err := ostack.NewImageClient(cloudProvider)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	data := []ImageData{}

	for _, v := range contents {
		var img *images.Image

		img, err = i.FetchImage(v)
		// If there was an error fetching the image we'll just move onto the next.
		if err != nil {
			log.Println(err)
			continue
		}

		// This captures when an image has been deleted but the scan data still exists.
		if img == nil {
			continue
		}

		data = append(data, ImageData{
			ImageID:    img.ID,
			ImageName:  img.Name,
			K8SVersion: img.Properties["k8s"],
			BuildDate:  img.Properties["date"],
			ScanStatus: img.Properties["security_scan"],
			TestStatus: img.Properties["e2e_test"],
		})
	}

	util.JSON(w, http.StatusOK, data)
}

// ApiV1GetImage will return both the scan and tests results along with some additional image data.
// This is especially useful when using a UI.
func (h *Handler) ApiV1GetImage(w http.ResponseWriter, r *http.Request, imageId generated.ImageID) {
	// Get image data from Openstack
	cloudProvider := ostack.NewCloudsProvider(h.cloudName)

	i, err := ostack.NewImageClient(cloudProvider)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	img, err := i.FetchImage(imageId)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	// Get Scan results
	scanRes, err := fetchScanData(h, imageId)
	if err != nil {
		util.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	//Get Tests results if dogkat is enabled
	var testRes generated.TestResult
	if h.dogkatS3 != nil {
		testRes, err = fetchTestData(h, imageId)
		if err != nil {
			// Just print an error here as we will want to continue with the scan data at the minimum
			log.Println(err)
			return
		}
	}

	imgData := ImageData{
		ImageID:    img.ID,
		ImageName:  img.Name,
		K8SVersion: img.Properties["k8s"],
		BuildDate:  img.Properties["date"],
		ScanStatus: img.Properties["security_scan"],
		ScanData:   scanRes,
		TestStatus: img.Properties["e2e_test"],
		TestData:   testRes,
	}

	util.JSON(w, http.StatusOK, imgData)
}
