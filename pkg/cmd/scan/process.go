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
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	sshconnect "github.com/eschercloudai/baski/pkg/ssh"
	"github.com/eschercloudai/baski/pkg/trivy"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/pkg/sftp"
	"log"
	"os"
	"time"
)

// FetchResultsFromServer pulls the results.json from the remote scanning server.
func FetchResultsFromServer(freeIP string, kp *keypairs.KeyPair) error {
	client, err := sshconnect.NewClient(kp, freeIP)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Println("Successfully connected to ssh server.")
	log.Println("waiting 2 minutes for the results of the scan to become available.")
	time.Sleep(2 * time.Minute)

	// open an SFTP session over an existing ssh connection.
	log.Println("pulling results.")
	sftpConnection, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpConnection.Close()

	resultsFile, err := sshconnect.CopyFromRemoteServer(sftpConnection, "/tmp/", "/tmp/", "results.json")
	if err != nil {
		log.Println(err.Error())
	}

	//Check there is data in the file
	fi, err := os.Stat(resultsFile.Name())
	if err != nil {
		log.Println(err.Error())
	}

	for fi.Size() == 0 {
		resultsFile, err = sshconnect.CopyFromRemoteServer(sftpConnection, "/tmp/", "/tmp/", "results.json")

		if err != nil {
			log.Println(err.Error())
		}

		fi, err = resultsFile.Stat()
		if err != nil {
			log.Println(err.Error())
		}
		time.Sleep(10 * time.Second)
	}

	resultsFile.Close()
	return err
}

// RemoveScanningResources cleans up the server and keypair from Openstack to ensure nothing is left lying around.
func RemoveScanningResources(serverID, keyName string, os *ostack.Client) {
	os.RemoveServer(serverID)
	os.RemoveKeypair(keyName)
}

// CheckForVulnerabilities will scan the results file for any 7+ CVE scores and if so will delete the image from Openstack and bail out here.
func CheckForVulnerabilities(checkScore float64, checkSeverity string) []trivy.Vulnerabilities {
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

	var vf []trivy.Vulnerabilities

	for _, r := range report.Results {
		for _, v := range r.Vulnerabilities {
			if checkSeverityThresholdPassed(v.Severity, v.Cvss, checkScore, checkSeverity) {
				vuln := trivy.Vulnerabilities{
					VulnerabilityID:  v.VulnerabilityID,
					PkgName:          v.PkgName,
					InstalledVersion: v.InstalledVersion,
					FixedVersion:     v.FixedVersion,
					Severity:         v.Severity,
					PublishedDate:    v.PublishedDate,
					LastModifiedDate: v.LastModifiedDate,
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
