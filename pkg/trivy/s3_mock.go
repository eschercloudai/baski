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

package trivy

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
)

type S3Mocked struct {
	mock.Mock
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

// FetchFromS3 checks the filename and return a cve value as a byte array as though it is returning the contents of an S3 pull.
func (s *S3Mocked) FetchFromS3(filename string) ([]byte, error) {
	if filename != "results_test.json" {
		return nil, errors.New("expected results_test.json for a filename")
	}
	return []byte("CVE-1234-56789"), nil
}

// PutToS3 checks the values that are being inputted to ensure it can run the actual command as it should
func (s *S3Mocked) PutToS3(contentType, key, fileName string, body io.ReadSeeker) error {
	expectedContentType := "text/plain"
	expectedKey := fmt.Sprintf("scans/%s/%s", "123456", "results.json")
	expectedFilename := "results.json"

	if contentType != expectedContentType {
		return fmt.Errorf("expected %s, got %s", expectedContentType, contentType)
	}

	if key != expectedKey {
		return fmt.Errorf("expected %s, got %s", expectedKey, key)
	}

	if fileName != "results.json" {
		return fmt.Errorf("expected %s, got %s", expectedFilename, fileName)
	}

	return nil
}
