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
	"fmt"
	"github.com/eschercloudai/baski/pkg/constants"
	"github.com/eschercloudai/baski/pkg/s3"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGenerateUserData(t *testing.T) {
	s3Mock := &s3.S3Mocked{
		Mock:      mock.Mock{},
		Endpoint:  "endpoint",
		AccessKey: "access-key",
		SecretKey: "secret-key",
		Bucket:    "bucket",
	}

	ignoreFile := "results_test.json"
	ignoreList := []string{"CVE-ABC-56789", "CVE-DEF-56789", "CVE-GHI-56789"}

	// Define test cases
	testCases := []struct {
		name           string
		s3             *s3.S3Mocked
		ignoreFile     string
		ignoreList     []string
		expectedResult []byte
	}{
		{
			name:           "Test case 1: No ignore file and empty ignore list",
			s3:             s3Mock,
			ignoreFile:     "",
			ignoreList:     nil,
			expectedResult: generateTestCase(false, false),
		},
		{
			name:           "Test case 2: With ignore file and empty ignore list",
			s3:             s3Mock,
			ignoreFile:     ignoreFile,
			ignoreList:     nil,
			expectedResult: generateTestCase(true, false),
		},
		{
			name:           "Test case 3: No ignore file and with ignore list",
			s3:             s3Mock,
			ignoreFile:     "",
			ignoreList:     ignoreList,
			expectedResult: generateTestCase(false, true),
		},
		{
			name:           "Test case 4: With ignore file and with ignore list",
			s3:             s3Mock,
			ignoreFile:     ignoreFile,
			ignoreList:     ignoreList,
			expectedResult: generateTestCase(true, true),
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateUserData(tc.s3, tc.ignoreFile, tc.ignoreList)
			if string(result) != string(tc.expectedResult) {
				t.Errorf("Test case %s failed. Expected:\n%s\nGot:\n%s", tc.name, string(tc.expectedResult), string(result))
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

	runScanCommand := "sudo trivy rootfs --scanners vuln -f json -o /tmp/results.json /;"

	// Prepare .trivyignore file
	if ignoreFile && !ignoreList {
		trivyIgnoreFile = `
cat << EOF > /tmp/.trivyignore
CVE-1234-56789
CVE-A1B2-56789

EOF
`
		runScanCommand = "sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -f json -o /tmp/results.json /;"
	} else if !ignoreFile && ignoreList {
		trivyIgnoreFile = `
cat << EOF > /tmp/.trivyignore

CVE-ABC-56789
CVE-DEF-56789
CVE-GHI-56789
EOF
`
		runScanCommand = "sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -f json -o /tmp/results.json /;"
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
		runScanCommand = "sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -f json -o /tmp/results.json /;"
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
