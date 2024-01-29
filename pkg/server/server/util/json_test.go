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

package util

import (
	"errors"
	"fmt"
	th "github.com/eschercloudai/baski/testhelpers"
	"io"
	"net/http"
	"testing"
)

// TestJSON generates a web response in json
func TestJSON(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()
	expected := `[{"Test":"some string","Value":1},{"Test":"another string","Value":2}]`

	th.Mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		code := http.StatusOK

		data := []struct {
			Test  string
			Value int
		}{
			{
				Test:  "some string",
				Value: 1,
			},
			{
				Test:  "another string",
				Value: 2,
			},
		}

		JSON(w, code, data)
	})

	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/test", th.Port))
	if err != nil {
		t.Error(err)
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if string(r) != expected {
		t.Errorf("expected %s, got %s", expected, r)
	}
}

// TestERROR generates a response in json from an error
func TestERROR(t *testing.T) {
	th.SetupPersistentPortHTTP(t, th.Port)
	defer th.TeardownHTTP()
	newError := errors.New("there was a problem")
	expected := `{"error":"there was a problem"}`

	th.Mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		code := http.StatusBadRequest
		ERROR(w, code, newError)
	})

	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/error", th.Port))
	if err != nil {
		t.Error(err)
	}

	r, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if string(r) != expected {
		t.Errorf("expected %s, got %s - %d", expected, r, []byte{10})
	}
}
