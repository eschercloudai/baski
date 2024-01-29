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

package ostack

import (
	"fmt"
	th "github.com/drewbernetes/baski/testhelpers"
	"net/http"
	"testing"
)

func TestNewComputeClient(t *testing.T) {
	//TODO: Implement this.
	// Not sure of the best approach for this yet.
}

// TestCreateKeypair creates a new KeyPair in Openstack.
func TestCreateKeypair(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/os-keypairs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, err := fmt.Fprint(w, th.CreateKeyPairOutput)
		if err != nil {
			t.Error(err)
			return
		}
	})

	cc := &ComputeClient{
		client: th.ServiceClient(),
	}
	key, err := cc.CreateKeypair("test-key")
	if err != nil {
		t.Error(err)
		return
	}
	if key.Name != "test-key" {
		t.Errorf("Expected %s, got %s", "test-key", key.Name)
	}
}

// TestRemoveKeypair will delete a Keypair from Openstack.
func TestRemoveKeypair(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/os-keypairs/test-key", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	cc := &ComputeClient{
		client: th.ServiceClient(),
	}
	err := cc.RemoveKeypair("test-key")
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
}

// TestCreateServer creates a compute instance in Openstack.
func TestCreateServer(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()
	th.Mux.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Header().Add("Content-Type", "application/json")
		_, err := fmt.Fprint(w, th.SingleServerBody)
		if err != nil {
			t.Error(err)
			return
		}
	})

	th.Mux.HandleFunc("/images/detail", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		err := r.ParseForm()
		if err != nil {
			t.Error(err)
			return
		}
		marker := r.Form.Get("marker")
		switch marker {
		case "":
			_, err = fmt.Fprint(w, `
				{
					"images": [
						{
							"status": "ACTIVE",
							"updated": "2014-09-23T12:54:56Z",
							"id": "f3e4a95d-1f4f-4989-97ce-f3a1fb8c04d7",
							"OS-EXT-IMG-SIZE:size": 476704768,
							"name": "F17-x86_64-cfntools",
							"created": "2014-09-23T12:54:52Z",
							"minDisk": 0,
							"progress": 100,
							"minRam": 0
						},
						{
							"status": "ACTIVE",
							"updated": "2014-09-23T12:51:43Z",
							"id": "f90f6034-2570-4974-8351-6b49732ef2eb",
							"OS-EXT-IMG-SIZE:size": 13167616,
							"name": "cirros-0.3.2-x86_64-disk",
							"created": "2014-09-23T12:51:42Z",
							"minDisk": 0,
							"progress": 100,
							"minRam": 0
						}
					]
				}
			`)
			if err != nil {
				t.Error(err)
				return
			}
		case "2":
			_, err = fmt.Fprint(w, `{ "images": [] }`)
			if err != nil {
				t.Error(err)
				return
			}
		default:
			t.Fatalf("Unexpected marker: [%s]", marker)
		}
	})

	th.Mux.HandleFunc("/flavors/detail", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		err := r.ParseForm()
		if err != nil {
			t.Error(err)
			return
		}
		marker := r.Form.Get("marker")
		switch marker {
		case "":
			_, err = fmt.Fprintf(w, `
						{
							"flavors": [
								{
									"id": "1",
									"name": "m1.tiny",
									"disk": 1,
									"ram": 512,
									"vcpus": 1,
									"swap":""
								},
								{
									"id": "2",
									"name": "m2.small",
									"disk": 10,
									"ram": 1024,
									"vcpus": 2,
									"swap": 1000
								}
							],
							"flavors_links": [
								{
									"href": "%s/flavors/detail?marker=2",
									"rel": "next"
								}
							]
						}
					`, th.Server.URL)
			if err != nil {
				t.Error(err)
			}
		case "2":
			_, err = fmt.Fprint(w, `{ "flavors": [] }`)
			if err != nil {
				t.Error(err)
			}
		default:
			t.Fatalf("Unexpected marker: [%s]", marker)
		}
	})

	cc := &ComputeClient{
		client: th.ServiceClient(),
	}
	configDrive := false
	s, err := cc.CreateServer("test-key", "1", "d32019d3-bc6e-4319-9c1d-6722fc136a22", &configDrive, []byte{}, "f90f6034-2570-4974-8351-6b49732ef2eb")
	if err != nil {
		t.Error(err)
		return
	}
	if s.ID != "9e5476bd-a4ec-4653-93d6-72c93aa682ba" {
		t.Errorf("Expected %s, got %s", "9e5476bd-a4ec-4653-93d6-72c93aa682ba", s.ID)
		return
	}
}

// TestGetServerStatus gets the status of a server
func TestGetServerStatus(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/servers/9e5476bd-a4ec-4653-93d6-72c93aa682ba", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, th.SingleServerBody)
		if err != nil {
			t.Error(err)
			return
		}
	})

	cc := &ComputeClient{
		client: th.ServiceClient(),
	}
	active, err := cc.GetServerStatus("9e5476bd-a4ec-4653-93d6-72c93aa682ba")
	if err != nil {
		t.Error(err)
		return
	}
	if !active {
		t.Errorf("Expected %t, got %t", true, active)
		return
	}
}

// TestAttachIP attaches the provided IP to the provided server.
func TestAttachIP(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()
	th.Mux.HandleFunc("/servers/9e5476bd-a4ec-4653-93d6-72c93aa682ba/action", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	cc := &ComputeClient{
		client: th.ServiceClient(),
	}
	err := cc.AttachIP("9e5476bd-a4ec-4653-93d6-72c93aa682ba", "2f245a7b-796b-4f26-9cf9-9e82d248fda7")
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
}

// TestRemoveServer will delete a Server from Openstack.
func TestRemoveServer(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/servers/9e5476bd-a4ec-4653-93d6-72c93aa682ba", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	cc := &ComputeClient{
		client: th.ServiceClient(),
	}
	err := cc.RemoveServer("9e5476bd-a4ec-4653-93d6-72c93aa682ba")
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
}

// TestGetFlavorIDByName will take a name of a flavor and attempt to find the ID from Openstack.
func TestGetFlavorIDByName(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/flavors/detail", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		err := r.ParseForm()
		if err != nil {
			t.Error(err)
			return
		}
		marker := r.Form.Get("marker")
		switch marker {
		case "":
			_, err = fmt.Fprintf(w, `
					{
						"flavors": [
							{
								"id": "1",
								"name": "m1.tiny",
								"vcpus": 1,
								"disk": 1,
								"ram": 9216000,
								"swap":"",
								"os-flavor-access:is_public": true,
								"OS-FLV-EXT-DATA:ephemeral": 10,
								"description": "foo"
							},
							{
								"id": "2",
								"name": "m1.small",
								"vcpus": 1,
								"disk": 20,
								"ram": 2048,
								"swap": 1000,
								"os-flavor-access:is_public": true,
								"OS-FLV-EXT-DATA:ephemeral": 0
							},
							{
								"id": "3",
								"name": "m1.medium",
								"vcpus": 2,
								"disk": 40,
								"ram": 4096,
								"swap": 1000,
								"os-flavor-access:is_public": false,
								"OS-FLV-EXT-DATA:ephemeral": 0
							}
						],
						"flavors_links": [
							{
								"href": "%s/flavors/detail?marker=2",
								"rel": "next"
							}
						]
					}
				`, th.Server.URL)
			if err != nil {
				t.Error(err)
			}
		case "2":
			_, err = fmt.Fprint(w, `{ "flavors": [] }`)
			if err != nil {
				t.Error(err)
			}
		default:
			t.Fatalf("Unexpected marker: [%s]", marker)
		}
	})

	cc := &ComputeClient{
		client: th.ServiceClient(),
	}
	fid, err := cc.GetFlavorIDByName("m1.tiny")
	if err != nil {
		t.Error(err)
		return
	}
	if fid != "1" {
		t.Errorf("Expected %s, got %s", "1", fid)
		return
	}
}
