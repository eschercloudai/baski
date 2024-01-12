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
	"bytes"
	"fmt"
	"github.com/eschercloudai/baski/pkg/constants"
	"github.com/eschercloudai/baski/pkg/mock"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		filename       string
		path           string
		ignoreList     []string
		severity       Severity
		expectedResult *TrivyOptions
	}{
		{
			name:       "Test case 1: Test ignorefile",
			filename:   ".trivyignore",
			path:       "",
			ignoreList: []string{},
			severity:   MEDIUM,
			expectedResult: &TrivyOptions{
				ignorePath:     "",
				ignoreFilename: ".trivyignore",
				ignoreList:     []string{},
				severity:       MEDIUM,
			},
		},
		{
			name:       "Test case 2: Test ignorefile with path",
			filename:   ".trivyignore",
			path:       "something",
			ignoreList: []string{},
			severity:   MEDIUM,
			expectedResult: &TrivyOptions{
				ignorePath:     "something",
				ignoreFilename: ".trivyignore",
				ignoreList:     []string{},
				severity:       MEDIUM,
			},
		},
		{
			name:       "Test case 3: Test ignorefile and list",
			filename:   ".trivyignore",
			path:       "",
			ignoreList: []string{"a", "b", "c"},
			severity:   HIGH,
			expectedResult: &TrivyOptions{
				ignorePath:     "",
				ignoreFilename: ".trivyignore",
				ignoreList:     []string{"a", "b", "c"},
				severity:       HIGH,
			},
		},
	}

	// RunScan the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			to := New(tc.path, tc.filename, nil, MEDIUM)

			if reflect.DeepEqual(to, tc.expectedResult) {
				t.Errorf("Expected data %v, got: %v\n", tc.expectedResult, to)
			}
		})
	}
}

func TestGetFilename(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		filename       string
		path           string
		expectedResult string
	}{
		{
			name:           "Test case 1: Test filename without path",
			filename:       ".trivyignore",
			path:           "",
			expectedResult: ".trivyignore",
		},
		{
			name:           "Test case 2: Test filename with path",
			filename:       ".trivyignore",
			path:           "something",
			expectedResult: "something/.trivyignore",
		},
	}

	// RunScan the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			to := New(tc.path, tc.filename, nil, MEDIUM)
			result := to.GetFilename()

			if result != tc.expectedResult {
				t.Errorf("Expected data %s, got: %s\n", tc.expectedResult, result)
			}
		})
	}
}

func TestGenerateUserData(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	m := mock.NewMockS3Interface(c)
	// Set expectations for Fetch.

	ignoreFile := ".trivyignore"
	fileBytes := []byte("CVE-1234-56789\nCVE-A1B2-56789")
	ignoreList := []string{"CVE-ABC-56789", "CVE-DEF-56789", "CVE-GHI-56789"}

	m.EXPECT().Fetch(ignoreFile).Return(fileBytes, nil).AnyTimes()

	// Define test cases
	testCases := []struct {
		name           string
		s3             *mock.MockS3Interface
		ignoreFile     string
		ignoreList     []string
		expectedResult []byte
	}{
		{
			name:           "Test case 1: No ignore file and empty ignore list",
			s3:             m,
			ignoreFile:     "",
			ignoreList:     []string{"[]"},
			expectedResult: generateTestCase(false, false),
		},
		{
			name:           "Test case 2: With ignore file and empty ignore list",
			s3:             m,
			ignoreFile:     ignoreFile,
			ignoreList:     []string{"[]"},
			expectedResult: generateTestCase(true, false),
		},
		{
			name:           "Test case 3: No ignore file and with ignore list",
			s3:             m,
			ignoreFile:     "",
			ignoreList:     ignoreList,
			expectedResult: generateTestCase(false, true),
		},
		{
			name:           "Test case 4: With ignore file and with ignore list",
			s3:             m,
			ignoreFile:     ignoreFile,
			ignoreList:     ignoreList,
			expectedResult: generateTestCase(true, true),
		},
	}

	// RunScan the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			to := New("", tc.ignoreFile, tc.ignoreList, MEDIUM)
			result, err := to.GenerateTrivyCommand(tc.s3)
			if err != nil {
				t.Errorf("Expected error %v, got: %s\n", nil, err.Error())
			}

			if !bytes.Equal(result, tc.expectedResult) {
				t.Errorf("Expected data %s, got: %s\n", tc.expectedResult, result)
			}
		})
	}
}

// generateTestCase creates a test case base on the inputs supplied
func generateTestCase(ignoreFile bool, ignoreList bool) []byte {

	var trivyIgnoreFile string

	// Prepare trivy setup
	trivyUserData := fmt.Sprintf(`if ! type trivy >/dev/null 2>&1; then
	wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_Linux-64bit.tar.gz" | tar xzf -;
	wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_checksums.txt";
	chmod u+x ./trivy;
	mv ./trivy /usr/local/bin/trivy;
fi`, constants.TrivyVersion, constants.TrivyVersion, constants.TrivyVersion, constants.TrivyVersion)

	runScanCommand := "sudo trivy rootfs --scanners vuln -s MEDIUM,HIGH,CRITICAL -f json -o /tmp/results.json /;"

	// Prepare .trivyignore file
	if ignoreFile && !ignoreList {
		trivyIgnoreFile = `
cat << EOF > /tmp/.trivyignore
CVE-1234-56789
CVE-A1B2-56789
EOF
`
		runScanCommand = "sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -s MEDIUM,HIGH,CRITICAL -f json -o /tmp/results.json /;"
	} else if !ignoreFile && ignoreList {
		trivyIgnoreFile = `
cat << EOF > /tmp/.trivyignore

CVE-ABC-56789
CVE-DEF-56789
CVE-GHI-56789
EOF
`
		runScanCommand = "sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -s MEDIUM,HIGH,CRITICAL -f json -o /tmp/results.json /;"
	} else if ignoreFile && ignoreList {
		trivyIgnoreFile = `
cat << EOF > /tmp/.trivyignore
CVE-1234-56789
CVE-A1B2-56789
CVE-ABC-56789
CVE-DEF-56789
CVE-GHI-56789
EOF
`
		runScanCommand = "sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -s MEDIUM,HIGH,CRITICAL -f json -o /tmp/results.json /;"
	}

	// Put it all together
	return []byte(fmt.Sprintf(`#!/bin/bash
touch /tmp/finished;
%s
%s
%s
echo done > /tmp/finished;
`, trivyIgnoreFile, trivyUserData, runScanCommand))
}
