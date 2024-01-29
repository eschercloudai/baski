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

func TestNewNetworkClient(t *testing.T) {
	//TODO: Implement this.
	// Not sure of the best approach for this yet.
}

// TestGetFloatingIP will create a new FIP.
func TestGetFloatingIP(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v2.0/networks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := fmt.Fprint(w, th.ListResponse)
		if err != nil {
			t.Error(err)
			return
		}
	})

	th.Mux.HandleFunc("/v2.0/floatingips", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		_, err := fmt.Fprint(w, th.FIPCreate)
		if err != nil {
			t.Error(err)
			return
		}
	})

	nc := &NetworkClient{
		client: th.CommonServiceClient(),
	}

	fip, err := nc.GetFloatingIP("public")
	if err != nil {
		t.Error(err)
		return
	}

	if fip.FloatingIP != "1.2.3.4" {
		t.Errorf(" expected %s, Got %s", "1.2.3.4", fip.FloatingIP)
	}
}

// TestRemoveFIP will delete a Floating IP from Openstack.
func TestRemoveFIP(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v2.0/floatingips/2f245a7b-796b-4f26-9cf9-9e82d248fda7", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	nc := &NetworkClient{
		client: th.CommonServiceClient(),
	}
	err := nc.RemoveFIP("2f245a7b-796b-4f26-9cf9-9e82d248fda7")
	if err != nil {
		t.Errorf("expected nil error, got %s", err.Error())
	}
}
