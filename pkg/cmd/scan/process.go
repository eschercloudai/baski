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

package scan

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eschercloudai/baski/pkg/cmd/util/data"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/eschercloudai/baski/pkg/s3"
	sshconnect "github.com/eschercloudai/baski/pkg/ssh"
	"github.com/eschercloudai/baski/pkg/trivy"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/pkg/sftp"
	"log"
	"os"
	"strings"
	"time"
)

// fetchResultsFromServer pulls the results.json from the remote scanning server.
func fetchResultsFromServer(freeIP string, kp *keypairs.KeyPair) error {
	client, err := sshconnect.NewClient(kp, freeIP)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Println("Successfully connected to ssh server")
	log.Println("waiting 4 minutes for Trivy to update and the results of the scan to become available")

	// open an SFTP session over an existing ssh connection.
	sftpConnection, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpConnection.Close()

	log.Println("checking for scan completion")
	retries := 20
	for !hasScanCompleted(sftpConnection) {
		if retries <= 0 {
			return errors.New("couldn't fetch the results - timed out waiting for condition")
		}
		log.Printf("scan still running... %d retries left\n", retries)
		time.Sleep(10 * time.Second)
		retries -= 1
	}
	log.Println("scan completed, fetching results")
	if err != nil {
		return err
	}
	defer sftpConnection.Close()

	_, err = sshconnect.CopyFromRemoteServer(sftpConnection, "/tmp/", "/tmp/", "results.json")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func hasScanCompleted(sftpConnection *sftp.Client) bool {
	status, err := sshconnect.CopyFromRemoteServer(sftpConnection, "/tmp/", "/tmp/", "finished")
	if err != nil {
		return false
	}

	fi, err := os.Stat(status.Name())
	if err != nil {
		log.Println(err.Error())
		return false
	}

	if fi.Size() == 0 {
		return false
	}
	return true
}

// removeScanningResources cleans up the server and keypair from Openstack to ensure nothing is left lying around.
func removeScanningResources(serverID, keyName string, fip *floatingips.FloatingIP, os *ostack.Client) {
	os.RemoveServer(serverID)
	os.RemoveKeypair(keyName)
	os.RemoveFIP(fip)
}

// checkForVulnerabilities will scan the results file for any 7+ CVE scores and if so will delete the image from Openstack and bail out here.
func checkForVulnerabilities(checkScore float64, checkSeverity string) []trivy.ScanFailedReport {
	log.Println("checking results for failures")
	rf, err := os.ReadFile("/tmp/results.json")
	if err != nil {
		log.Println(err)
		return nil
	}

	report := &trivy.Report{}

	err = json.Unmarshal(rf, report)

	if err != nil {
		log.Println(err)
		return nil
	}

	var vf []trivy.ScanFailedReport

	for _, r := range report.Results {
		for _, v := range r.Vulnerabilities {
			if checkSeverityThresholdPassed(v.Severity, v.Cvss, checkScore, checkSeverity) {
				vuln := trivy.ScanFailedReport{
					VulnerabilityID:  v.VulnerabilityID,
					Description:      v.Description,
					PkgName:          v.PkgName,
					InstalledVersion: v.InstalledVersion,
					FixedVersion:     v.FixedVersion,
					Severity:         v.Severity,
				}
				// We don't need all scores in here, so we just grab the one that triggered the threshold
				if v.Cvss.Nvd != nil {
					if v.Cvss.Nvd.V3Score >= checkScore {
						vuln.Cvss = trivy.CVSS{Nvd: &trivy.Score{V3Score: v.Cvss.Nvd.V3Score}}
					} else if v.Cvss.Nvd.V2Score >= checkScore {
						vuln.Cvss = trivy.CVSS{Nvd: &trivy.Score{V2Score: v.Cvss.Nvd.V2Score}}
					}
				} else if v.Cvss.Redhat != nil {
					if v.Cvss.Redhat.V3Score >= checkScore {
						vuln.Cvss = trivy.CVSS{Redhat: &trivy.Score{V3Score: v.Cvss.Redhat.V3Score}}
					} else if v.Cvss.Redhat.V2Score >= checkScore {
						vuln.Cvss = trivy.CVSS{Redhat: &trivy.Score{V2Score: v.Cvss.Redhat.V2Score}}
					}
				} else if v.Cvss.Ghsa != nil {
					if v.Cvss.Ghsa.V3Score >= checkScore {
						vuln.Cvss = trivy.CVSS{Ghsa: &trivy.Score{V3Score: v.Cvss.Ghsa.V3Score}}
					}
				}

				vf = append(vf, vuln)
			}
		}
	}
	return vf
}

// checkSeverityThresholdPassed checks for a score that is >= checkScore and checkSeverity. It will return true if so.
func checkSeverityThresholdPassed(severity string, cvss trivy.CVSS, checkScore float64, checkSeverity string) bool {
	if cvss.Nvd != nil {
		if cvss.Nvd.V3Score >= checkScore && trivy.CheckSeverity(severity, checkSeverity) {
			return true
		} else if cvss.Nvd.V2Score >= checkScore && trivy.CheckSeverity(severity, checkSeverity) {
			return true
		}
	}
	if cvss.Redhat != nil {
		if cvss.Redhat.V3Score >= checkScore && trivy.CheckSeverity(severity, checkSeverity) {
			return true
		} else if cvss.Redhat.V2Score >= checkScore && trivy.CheckSeverity(severity, checkSeverity) {
			return true
		}
	}
	if cvss.Ghsa != nil {
		if cvss.Ghsa.V3Score >= checkScore && trivy.CheckSeverity(severity, checkSeverity) {
			return true
		}
	}
	return false
}

func runScan(osClient *ostack.Client, o *flags.ScanOptions, img *images.Image) error {

	//TODO: We need to capture-ctl c and cleanup resources if it's hit.

	kp, err := osClient.CreateKeypair(img.ID)
	if err != nil {
		return err
	}

	fip, err := osClient.GetFloatingIP(strings.ToLower(o.FloatingIPNetworkName))
	if err != nil {
		osClient.RemoveKeypair(kp.Name)
		return err
	}

	s3Connection := &s3.S3{
		Endpoint:  o.Endpoint,
		AccessKey: o.AccessKey,
		SecretKey: o.SecretKey,
		Bucket:    o.ScanBucket,
	}

	s3Path := fmt.Sprintf("%s/%s", o.TrivyignorePath, o.TrivyignoreFilename)
	if o.TrivyignorePath == "" {
		s3Path = o.TrivyignoreFilename
	}

	userData := trivy.GenerateUserData(s3Connection, s3Path, o.TrivyignoreList)

	server, err := osClient.CreateServer(kp, o, userData, img.ID)
	if err != nil {
		osClient.RemoveKeypair(kp.Name)
		osClient.RemoveFIP(fip)
		return err
	}

	state := osClient.GetServerStatus(server.ID)
	checkLimit := 0
	for !state {
		if checkLimit > 100 {
			panic(errors.New("server failed to com online after 500 seconds"))
		}
		log.Println("server not active, waiting 5 seconds and then checking again")
		time.Sleep(5 * time.Second)
		state = osClient.GetServerStatus(server.ID)
		checkLimit++
	}

	err = osClient.AttachIP(server.ID, fip.IP)
	if err != nil {
		removeScanningResources(server.ID, kp.Name, fip, osClient)
		return err
	}

	err = fetchResultsFromServer(fip.IP, kp)
	if err != nil {
		removeScanningResources(server.ID, kp.Name, fip, osClient)
		return err
	}

	// Cleanup the scanning resources
	removeScanningResources(server.ID, kp.Name, fip, osClient)

	// Default to replace unless the field isn't found below
	operation := images.ReplaceOp

	field, err := data.GetNestedField(img.Properties, "image", "metadata", "security_scan")
	if err != nil || field == nil {
		operation = images.AddOp
	}
	metaValue := "passed"
	resultsFile := fmt.Sprintf("/tmp/%s.json", img.ID)

	scoreCheck := checkForVulnerabilities(o.MaxSeverityScore, strings.ToUpper(o.MaxSeverityType))
	if len(scoreCheck) != 0 {
		if o.AutoDeleteImage {
			osClient.RemoveImage(img.ID)
		}
		var j []byte
		j, err = json.Marshal(scoreCheck)
		if err != nil {
			return errors.New("couldn't marshall vulnerability trivyIgnoreFile: " + err.Error())
		}

		// write the vulnerabilities into the results file
		err = os.WriteFile(resultsFile, j, os.FileMode(0644))
		if err != nil {
			return errors.New("couldn't write vulnerability trivyIgnoreFile to file: " + err.Error())
		}

		metaValue = "failed"
	}
	_, err = osClient.ModifyImageMetadata(img.ID, "security_scan", metaValue, operation)
	if err != nil {
		return err
	}

	//Upload results to S3
	f, err := os.Open(resultsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	err = s3Connection.PutToS3("text/plain", fmt.Sprintf("scans/%s/%s", img.ID, "results.json"), "results.json", f)
	if err != nil {
		return err
	}

	if !o.SkipCVECheck {
		return errors.New("vulnerabilities detected above threshold - removed image from infra. Please see the possible fixes located at '/tmp/results.json' for further information on this")
	}
	return nil
}
