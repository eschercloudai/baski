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

package ostack

import (
	"fmt"
	"github.com/eschercloudai/baski/pkg/util/flags"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

var (
	testImages = []imageEntry{
		{
			ID: "cirros-0.3.4-x86_64-uec",
			JSON: `{
            "status": "active",
            "name": "cirros-0.3.4-x86_64-uec",
            "tags": [],
            "kernel_id": "e1b6edd4-bd9b-40ac-b010-8a6c16de4ba4",
            "container_format": "ami",
            "created_at": "2015-07-15T11:43:35Z",
            "ramdisk_id": "8c64f48a-45a3-4eaa-adff-a8106b6c005b",
            "disk_format": "ami",
            "updated_at": "2015-07-15T11:43:35Z",
            "visibility": "public",
            "self": "/v2/images/07aa21a9-fa1a-430e-9a33-185be5982431",
            "min_disk": 0,
            "protected": false,
            "id": "07aa21a9-fa1a-430e-9a33-185be5982431",
            "size": 25165824,
            "file": "/v2/images/07aa21a9-fa1a-430e-9a33-185be5982431/file",
            "checksum": "eb9139e4942121f22bbc2afc0400b2a4",
            "owner": "cba624273b8344e59dd1fd18685183b0",
            "virtual_size": null,
            "min_ram": 0,
            "schema": "/v2/schemas/image",
            "hw_disk_bus": "scsi",
            "hw_disk_bus_model": "virtio-scsi",
            "hw_scsi_model": "virtio-scsi"
        }`,
		},
		{
			ID: "cirros-0.3.4-x86_64-uec-ramdisk",
			JSON: `{
            "status": "active",
            "name": "cirros-0.3.4-x86_64-uec-ramdisk",
            "tags": [],
            "container_format": "ari",
            "created_at": "2015-07-15T11:43:32Z",
            "size": 3740163,
            "disk_format": "ari",
            "updated_at": "2015-07-15T11:43:32Z",
            "visibility": "public",
            "self": "/v2/images/8c64f48a-45a3-4eaa-adff-a8106b6c005b",
            "min_disk": 0,
            "protected": false,
            "id": "8c64f48a-45a3-4eaa-adff-a8106b6c005b",
            "file": "/v2/images/8c64f48a-45a3-4eaa-adff-a8106b6c005b/file",
            "checksum": "be575a2b939972276ef675752936977f",
            "owner": "cba624273b8344e59dd1fd18685183b0",
            "virtual_size": null,
            "min_ram": 0,
            "schema": "/v2/schemas/image",
            "hw_disk_bus": "scsi",
            "hw_disk_bus_model": "virtio-scsi",
            "hw_scsi_model": "virtio-scsi"
        }`,
		},
		{
			ID: "cirros-0.3.4-x86_64-uec-kernel",
			JSON: `{
            "status": "active",
            "name": "cirros-0.3.4-x86_64-uec-kernel",
            "tags": [],
            "container_format": "aki",
            "created_at": "2015-07-15T11:43:29Z",
            "size": 4979632,
            "disk_format": "aki",
            "updated_at": "2015-07-15T11:43:30Z",
            "visibility": "public",
            "self": "/v2/images/e1b6edd4-bd9b-40ac-b010-8a6c16de4ba4",
            "min_disk": 0,
            "protected": false,
            "id": "e1b6edd4-bd9b-40ac-b010-8a6c16de4ba4",
            "file": "/v2/images/e1b6edd4-bd9b-40ac-b010-8a6c16de4ba4/file",
            "checksum": "8a40c862b5735975d82605c1dd395796",
            "owner": "cba624273b8344e59dd1fd18685183b0",
            "virtual_size": null,
            "min_ram": 0,
            "schema": "/v2/schemas/image",
            "hw_disk_bus": "scsi",
            "hw_disk_bus_model": "virtio-scsi",
            "hw_scsi_model": "virtio-scsi"
        }`,
		},
	}
)

type imageEntry struct {
	ID   string
	JSON string
}

// TestNewImageClient
func TestNewImageClient(t *testing.T) {
	//TODO: Implement this.
	// Not sure of the best approach for this yet.
}

// TestGenerateBuilderMetadata generates some glance metadata for the image.
func TestGenerateBuilderMetadata(t *testing.T) {
	th.SetupPersistentPortHTTP(t, port)
	defer th.TeardownHTTP()

	tests := []struct {
		name     string
		options  *flags.BuildOptions
		expected map[string]string
	}{
		{
			name: "Test with GPU",
			options: &flags.BuildOptions{
				AddNvidiaSupport: true,
				NvidiaVersion:    "1.2.3",
				BuildOS:          "ubuntu",
				KubeVersion:      "1.28",
				OpenStackFlags: flags.OpenStackFlags{
					RootfsUUID: "123456",
				},
			},
			expected: map[string]string{
				"os":          "ubuntu",
				"k8s":         "1.28",
				"gpu":         "1.2.3",
				"date":        "2006-01-02T15:04:05Z07:00",
				"rootfs_uuid": "123456",
			},
		},
		{
			name: "Test without GPU",
			options: &flags.BuildOptions{
				AddNvidiaSupport: false,
				BuildOS:          "ubuntu",
				KubeVersion:      "1.28",
				OpenStackFlags: flags.OpenStackFlags{
					RootfsUUID: "123456",
				},
			},
			expected: map[string]string{
				"os":          "ubuntu",
				"k8s":         "1.28",
				"gpu":         "no_gpu",
				"date":        "2006-01-02T15:04:05Z07:00",
				"rootfs_uuid": "123456",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta := GenerateBuilderMetadata(tc.options)
			//We override the dat here as it's based off of time.Now()
			meta["date"] = "2006-01-02T15:04:05Z07:00"

			if !reflect.DeepEqual(meta, tc.expected) {
				t.Errorf("Expected %+v, got %+v", tc.expected, meta)
			}
		})
	}

}

// TestModifyImageMetadata allows image metadata to be added, updated or removed.
func TestModifyImageMetadata(t *testing.T) {
	th.SetupPersistentPortHTTP(t, port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/images/da3b75d9-3f4a-40e7-8a2c-bfab23927dea", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"id": "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
			"name": "Fedora 17",
			"status": "active",
			"visibility": "public",
			"os_hidden": false,
			"protected": true,
			"size": 2254249,
			"checksum": "2cec138d7dae2aa59038ef8c9aec2390",
			"tags": [
				"fedora",
				"beefy"
			],
			"created_at": "2012-08-10T19:23:50Z",
			"updated_at": "2012-08-12T11:11:33Z",
			"self": "/v2/images/da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
			"file": "/v2/images/da3b75d9-3f4a-40e7-8a2c-bfab23927dea/file",
			"schema": "/v2/schemas/image",
			"owner": "",
			"min_ram": 1024,
			"min_disk": 21,
			"disk_format": "",
			"virtual_size": 0,
			"container_format": "",
			"empty_value": "",
			"hw_disk_bus": "scsi",
			"hw_disk_bus_model": "virtio-scsi",
			"hw_scsi_model": "virtio-scsi",
			"test": "test-value"
		}`)
	})

	i := ImageClient{client: client.ServiceClient()}
	img, err := i.ModifyImageMetadata("da3b75d9-3f4a-40e7-8a2c-bfab23927dea", "test", "test-value", images.ReplaceOp)
	if err != nil {
		t.Error(err)
		return
	}
	if img.Properties["test"] != "test-value" {
		t.Errorf("Expected %s, got %+v", "test-value", img)
		return
	}
}

// TestRemoveImage will delete an image from Openstack.
func TestRemoveImage(t *testing.T) {
	th.SetupPersistentPortHTTP(t, port)
	defer th.TeardownHTTP()
	th.Mux.HandleFunc("/images/da3b75d9-3f4a-40e7-8a2c-bfab23927dea", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	i := ImageClient{client: client.ServiceClient()}
	err := i.RemoveImage("da3b75d9-3f4a-40e7-8a2c-bfab23927dea")
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
}

// TestFetchAllImages Fetches all the images from Openstack so that they can parsed after.
// // Because silly GopherCloud - or maybe OpenStack itself doesn't support wildcard search on names
// // and the tag search is limited to an id+tag :facepalm:
// // This probably can be improved though to prevent fetching billions of images.
func TestFetchAllImages(t *testing.T) {
	th.SetupPersistentPortHTTP(t, port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		limit := 10
		var err error
		if r.FormValue("limit") != "" {
			limit, err = strconv.Atoi(r.FormValue("limit"))
			if err != nil {
				t.Errorf("Error value for 'limit' parameter %v (error: %v)", r.FormValue("limit"), err)
			}

		}

		marker := ""
		newMarker := ""

		if r.Form["marker"] != nil {
			marker = r.Form["marker"][0]
		}

		t.Logf("limit = %v   marker = %v", limit, marker)

		selected := 0
		addNext := false
		var imageJSON []string

		_, err = fmt.Fprint(w, `{"images": [`)
		if err != nil {
			t.Error(err)
			return
		}

		for _, i := range testImages {
			if marker == "" || addNext {
				t.Logf("Adding image %v to page", i.ID)
				imageJSON = append(imageJSON, i.JSON)
				newMarker = i.ID
				selected++
			} else {
				if strings.Contains(i.JSON, marker) {
					addNext = true
				}
			}

			if selected == limit {
				break
			}
		}
		t.Logf("Writing out %v image(s)", len(imageJSON))
		_, err = fmt.Fprint(w, strings.Join(imageJSON, ","))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = fmt.Fprintf(w, `],
			    "next": "/images?marker=%s&limit=%v",
			    "schema": "/schemas/images",
			    "first": "/images?limit=%v"}`, newMarker, limit, limit)
		if err != nil {
			t.Error(err)
			return
		}

	})

	i := ImageClient{client: client.ServiceClient()}
	imgs, err := i.FetchAllImages("cirros-0.3.4-x86_64-uec")
	if err != nil {
		t.Error(err)
		return
	}

	expectedImage := testImages[0]
	var foundImage images.Image

	for _, v := range imgs {
		if v.Name == expectedImage.ID {
			foundImage = v
			break
		}
	}

	if expectedImage.ID != foundImage.Name {
		t.Errorf("Expected %+v, got %+v", expectedImage.ID, foundImage.Name)
		return
	}
}

// TestFetchImage allows us to fetch a single image by the id.
func TestFetchImage(t *testing.T) {
	th.SetupPersistentPortHTTP(t, port)
	defer th.TeardownHTTP()

	//TODO: Took this from gophercloud - not entirely sure it's a great test so may need refining.
	// But for now, it's a test.
	th.Mux.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		limit := 10
		var err error
		if r.FormValue("limit") != "" {
			limit, err = strconv.Atoi(r.FormValue("limit"))
			if err != nil {
				t.Errorf("Error value for 'limit' parameter %v (error: %v)", r.FormValue("limit"), err)
			}

		}

		marker := ""
		newMarker := ""

		if r.Form["marker"] != nil {
			marker = r.Form["marker"][0]
		}

		t.Logf("limit = %v   marker = %v", limit, marker)

		selected := 0
		addNext := false
		var imageJSON []string

		fmt.Fprintf(w, `{"images": [`)

		for _, i := range testImages {
			if marker == "" || addNext {
				t.Logf("Adding image %v to page", i.ID)
				imageJSON = append(imageJSON, i.JSON)
				newMarker = i.ID
				selected++
			} else {
				if strings.Contains(i.JSON, marker) {
					addNext = true
				}
			}

			if selected == limit {
				break
			}
		}
		t.Logf("Writing out %v image(s)", len(imageJSON))
		_, err = fmt.Fprint(w, strings.Join(imageJSON, ","))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = fmt.Fprintf(w, `],
			    "next": "/images?marker=%s&limit=%v",
			    "schema": "/schemas/images",
			    "first": "/images?limit=%v"}`, newMarker, limit, limit)
		if err != nil {
			t.Error(err)
			return
		}

	})

	i := ImageClient{client: client.ServiceClient()}
	resImg, err := i.FetchImage("8c64f48a-45a3-4eaa-adff-a8106b6c005b")
	if err != nil {
		t.Error(err)
		return
	}

	expectedImage := testImages[1]

	if expectedImage.ID != resImg.Name {
		t.Errorf("Expected %+v, got %+v", expectedImage.ID, resImg.Name)
		return
	}
}
