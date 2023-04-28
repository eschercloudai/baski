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
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"log"
	"os"
	"strings"

	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/eschercloudai/baski/pkg/trivy"
	"github.com/spf13/cobra"
)

// NewScanCommand creates a command that allows the scanning of an image.
func NewScanCommand() *cobra.Command {
	o := &flags.ScanOptions{}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan image",
		Long: `Scan image.

Scanning an image requires the creation of a new instance in Openstack using the image you want to scan.
Then, Trivy needs downloading and running against the filesystem. Again, this is time consuming.

The scan section of Baski fixes this for you and allows you to drink <enter drink here> whilst it does the hard work for you.

It does the following:
* Creates a new instance using the provided Openstack configuration variables
* Check if Trivy is available already, if not it'll download it
* Scans the rootfs
* Generates a report file that you can read with your eyes or via other means

If the checks for CVE flags/config values are set then it will bail out and generate a report with the CVEs that caused it to do so.
`,
		Run: func(cmd *cobra.Command, args []string) {
			o.SetOptionsFromViper()

			if !trivy.ValidSeverity(strings.ToUpper(o.MaxSeverityType)) {
				log.Fatalln("severity value passed is invalid. Allowed values are: NONE, LOW, MEDIUM, HIGH, CRITICAL")
			}

			cloudsConfig := ostack.InitOpenstack(o.CloudsPath)
			cloudsConfig.SetOpenstackEnvs(o.CloudName)

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[o.CloudName])

			kp := osClient.CreateKeypair(o.ImageID)
			server, freeIP := osClient.CreateServer(kp, o)

			err := FetchResultsFromServer(freeIP, kp)
			if err != nil {
				RemoveScanningResources(server.ID, kp.Name, osClient)
				log.Fatalln(err.Error())
			}
			if !o.SkipCVECheck {
				scoreCheck := CheckForVulnerabilities(o.MaxSeverityScore, strings.ToUpper(o.MaxSeverityType))
				if len(scoreCheck) != 0 {
					// Cleanup the scanning resources
					RemoveScanningResources(server.ID, kp.Name, osClient)

					if o.AutoDeleteImage {
						osClient.RemoveImage(o.ImageID)
					}

					var j []byte
					j, err = json.Marshal(scoreCheck)
					if err != nil {
						log.Fatalln("couldn't marshall vulnerability data")
					}

					// empty out the results json - we don't need the original since threshold vulnerabilities were found.
					err = os.Truncate("/tmp/results.json", 0)
					if err != nil {
						log.Fatalln("couldn't empty the results file")
					}

					// write the vulnerabilities into the results file
					err = os.WriteFile("/tmp/results.json", j, os.FileMode(0644))
					if err != nil {
						log.Fatalln("couldn't write vulnerability data to file")
					}

					var scanMsg string
					if o.AutoDeleteImage {
						scanMsg = "Vulnerabilities detected above threshold - removed image from infra. Please see the possible fixes located at '/tmp/results.json' for further information on this."
					} else {
						scanMsg = "Vulnerabilities detected above threshold - the image has not been removed from infra. Please see the possible fixes located at '/tmp/results.json' for further information on this."
					}
					log.Fatalln(scanMsg)
				}
			}

			// Cleanup the scanning resources
			RemoveScanningResources(server.ID, kp.Name, osClient)
		},
	}

	o.AddFlags(cmd)

	return cmd
}
