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
	th "github.com/eschercloudai/baski/testhelpers"
	"net/http"
	"reflect"
	"testing"
)

// TestInitOpenstack will read the contents of the clouds.yaml file for Openstack and parse it into a OpenstackClouds struct.
func TestNewCloudsProvider(t *testing.T) {
	account := "test-account"
	c := NewCloudsProvider(account)
	expected := CloudsProvider{cloud: account}

	if reflect.DeepEqual(&c, expected) {
		t.Errorf("Expected %+v, Got %+v", expected, c)
	}
}

// NewOpenstackClient creates the initial client for connecting to Openstack.
func TestClient(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
			{
				"versions": {
					"values": [
						{
							"status": "stable",
							"id": "v3.0",
							"links": [
								{ "href": "%s/v3.0", "rel": "self" }
							]
						},
						{
							"status": "stable",
							"id": "v2.0",
							"links": [
								{ "href": "%s/v2.0", "rel": "self" }
							]
						}
					]
				}
			}
		`, th.Server.URL, th.Server.URL)
	})

	th.Mux.HandleFunc("/v3.0/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Subject-Token", th.ID)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{ "token": { "expires_at": "2023-12-12T12:59:59.000000Z" } }`)
	})

	err := th.GenerateCloudsFile()
	if err != nil {
		t.Error(err)
		return
	}

	p := &CloudsProvider{cloud: "test-account"}
	_, err = p.Client()
	if err != nil {
		t.Error(err)
		return
	}
}
