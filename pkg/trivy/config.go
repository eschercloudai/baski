package trivy

import (
	"fmt"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/eschercloudai/baski/pkg/constants"
	"github.com/rhnvrm/simples3"
	"io"
	"log"
)

// GenerateTrivyFile generates the trivyignore file to be used during the scan.
func GenerateTrivyFile(o *flags.ScanOptions) []byte {
	var ignoreListData, trivyIgnoreFile []byte
	var err error

	// Check if a list of CVEs was passed in before checking for a trivyIgnore file
	if len(o.TrivyignoreList) != 0 {
		ignoreListData = parseIgnoreList(o.TrivyignoreList)
	}

	if len(o.TrivyignoreFilename) != 0 {
		trivyIgnoreFile, err = fetchTrivyFileFromS3(o.Endpoint, o.AccessKey, o.SecretKey, o.TrivyignoreBucket, o.TrivyignoreFilename)
		if err != nil {
			log.Printf("error: %s\n", err)
		}
	}

	return []byte(fmt.Sprintf("%s %s", string(trivyIgnoreFile), string(ignoreListData)))
}

// GenerateUserData Creates the user data that will be passed to the server being created so that a .trivyignore can be added and the scan can be run as per the users wishes.
func GenerateUserData(trivyIgnoreData []byte) []byte {
	log.Println("generating userdata")

	var trivyIgnoreFile string

	// Prepare trivy setup
	trivyUserData := fmt.Sprintf(`if ! type trivy >/dev/null 2>&1; then
	wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_Linux-64bit.tar.gz" | tar xzf -;
	wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_checksums.txt";
	chmod u+x ./trivy;
	mv ./trivy /usr/local/bin/trivy;
fi`, constants.TrivyVersion, constants.TrivyVersion, constants.TrivyVersion, constants.TrivyVersion)

	runScanData := "sudo trivy rootfs --scanners vuln -f json -o /tmp/results.json /;"

	// Prepare .trivyignore file
	if len(trivyIgnoreData) > 0 {
		trivyIgnoreFile = fmt.Sprintf(`
cat << EOF > /tmp/.trivyignore
%s
EOF
`, string(trivyIgnoreData))

		runScanData = "sudo trivy rootfs --ignorefile /tmp/.trivyignore --scanners vuln -f json -o /tmp/results.json /;"
	}

	// Put it all together
	return []byte(fmt.Sprintf(`#!/bin/bash
touch /tmp/finished;
%s
%s
%s
echo done > /tmp/finished;
`, trivyIgnoreFile, trivyUserData, runScanData))

}

// parseIgnoreList turns the ignore list passed into a format that can be used in the trivyignore file.
func parseIgnoreList(ignoreList []string) []byte {
	var list string

	for i := 0; i < len(ignoreList); i++ {
		list = fmt.Sprintf("\n%s\n%s\n", list, ignoreList[i])
	}

	return []byte(list)
}

// fetchTrivyFileFromS3 Downloads the trivyignore file from an S3 bucket and returns its contents as a byte array.
func fetchTrivyFileFromS3(endpoint string, accessKey string, secretKey string, bucket string, key string) ([]byte, error) {
	s3 := simples3.New("us-east-1", accessKey, secretKey)
	s3.SetEndpoint(endpoint)

	// Download the file.
	file, err := s3.FileDownload(simples3.DownloadInput{
		Bucket:    bucket,
		ObjectKey: key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %v", err)
	}

	return data, nil
}
