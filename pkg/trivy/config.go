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
	"github.com/eschercloudai/baski/pkg/util"
	"log"
	"strings"
)

type TrivyOptions struct {
	ignorePath     string
	ignoreFilename string
	ignoreList     []string
	severity       Severity
}

func New(filePath, filename string, ignoreList []string, severity Severity) *TrivyOptions {
	return &TrivyOptions{
		ignorePath:     filePath,
		ignoreFilename: filename,
		ignoreList:     ignoreList,
		severity:       severity,
	}
}

func (t *TrivyOptions) GetFilename() string {
	filename := t.ignoreFilename
	if t.ignoreFilename != "" {
		if t.ignorePath != "" {
			filename = fmt.Sprintf("%s/%s", t.ignorePath, t.ignoreFilename)
		}
	}
	return filename
}

// GenerateTrivyCommand Creates the user data that will be passed to the server being created so that a .trivyignore can be added and the scan can be run as per the users wishes.
func (t *TrivyOptions) GenerateTrivyCommand(s3 util.S3Interface) ([]byte, error) {
	trivyIgnoreData := generateTrivyFile(s3, t.GetFilename(), t.ignoreList)

	log.Println("generating userdata")

	var trivyIgnoreFile string

	// Prepare trivy setup
	trivyUserData := fmt.Sprintf(`if ! type trivy >/dev/null 2>&1; then
	wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_Linux-64bit.tar.gz" | tar xzf -;
	wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_checksums.txt";
	chmod u+x ./trivy;
	mv ./trivy /usr/local/bin/trivy;
fi`, constants.TrivyVersion, constants.TrivyVersion, constants.TrivyVersion, constants.TrivyVersion)

	if !ValidSeverity(t.severity) {
		return nil, fmt.Errorf("severity value passed is invalid. Allowed values are: UNKNOWN, LOW, MEDIUM, HIGH, CRITICAL")
	}

	severity := ParseSeverity(t.severity)
	severityList := strings.Join(severity, ",")

	// Set the default command to run here - it may get overridden later.
	runScanCommand := fmt.Sprintf("sudo trivy rootfs --scanners vuln -s %s -f json -o /tmp/results.json /;", severityList)

	// Prepare .trivyignore file
	if len(trivyIgnoreData) > 0 {
		trivyIgnoreFile = fmt.Sprintf(`
cat << EOF > /tmp/.trivyignore
%s
EOF
`, string(trivyIgnoreData))

		//Override the command to run as we now have a .trivyignore to add
		runScanCommand = fmt.Sprintf("sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -s %s -f json -o /tmp/results.json /;", severityList)
	}

	// Put it all together
	return []byte(fmt.Sprintf(`#!/bin/bash
touch /tmp/finished;
%s
%s
%s
echo done > /tmp/finished;
`, trivyIgnoreFile, trivyUserData, runScanCommand)), nil

}

// generateTrivyFile generates the trivyignore file to be used during the scan.
func generateTrivyFile(s3 util.S3Interface, ignoreFileName string, ignoreList []string) []byte {
	var ignoreListData, trivyIgnoreFile []byte
	var err error

	//We return nothing if there are no checks required
	if ignoreList[0] == "[]" && len(ignoreFileName) == 0 {
		return nil
	}

	// Check if a list of CVEs was passed in before checking for a trivyIgnore file
	if ignoreList[0] != "[]" {
		ignoreListData = parseIgnoreList(ignoreList)
	}

	if len(ignoreFileName) != 0 {
		trivyIgnoreFile, err = s3.Fetch(ignoreFileName)
		if err != nil {
			log.Printf("error: %s\n", err)
		}
	}

	data := trivyIgnoreFile

	if ignoreListData != nil {
		data = []byte(fmt.Sprintf("%s\n%s", string(trivyIgnoreFile), string(ignoreListData)))
	}
	return data
}

// parseIgnoreList turns the ignore list passed into a format that can be used in the trivyignore file.
func parseIgnoreList(ignoreList []string) []byte {
	list := strings.Join(ignoreList, "\n")

	return []byte(list)
}
