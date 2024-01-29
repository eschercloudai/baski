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

package server

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	opts := Options{
		ListenAddress: "127.0.0.1",
		ListenPort:    5000,
		Endpoint:      "",
		AccessKey:     "",
		SecretKey:     "",
		Bucket:        "",
	}
	server := Server{Options: opts}
	s, err := server.NewServer(true)
	if err != nil {
		t.Error(err)
	}
	go func() {
		if err = s.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatalln(err, "unexpected server error")
			return
		}
	}()

	// Just sleep a second to ensure the server is started
	time.Sleep(1 * time.Second)

	res, err := http.Get("http://127.0.0.1:5000/healthz")
	if err != nil {
		t.Error(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	expected := "ok"

	if string(body) != expected {
		t.Errorf("expected %s, got %s", expected, string(body))
	}

}

// CORSAllowOriginAllMiddleware sets the header for Access-Control-Allow-Origin = "*"
func TestCORSAllowOriginAllMiddleware(t *testing.T) {
	// Create a test HTTP handler function that simulates a request
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Test Response"))
		if err != nil {
			t.Error(err)
		}
	}

	// Create a request to pass through the middleware
	req := httptest.NewRequest("GET", "/", nil)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the middleware with the test handler
	middleware := CORSAllowOriginAllMiddleware(handler)
	middleware(rr, req)

	// Check if the headers have been set correctly
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, X-CSRF-Token",
		"Access-Control-Allow-Credentials": "true",
	}

	for key, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("Header %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Check if the response body is as expected
	if rr.Body.String() != "Test Response" {
		t.Errorf("Expected response body: 'Test Response', got: '%s'", rr.Body.String())
	}
}

// OptionsPreflightAllow sets the header for Access-Control-Allow-Origin = "*"
func TestOptionsPreflightAllow(t *testing.T) {
	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the OptionsPreflightAllow function
	responseWriter := OptionsPreflightAllow(rr)

	// Check if the headers have been set correctly
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, X-CSRF-Token",
		"Access-Control-Allow-Credentials": "true",
	}

	for key, expectedValue := range expectedHeaders {
		actualValue := responseWriter.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("Header %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}
