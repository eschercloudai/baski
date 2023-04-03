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
	"log"
	"os"
	"strings"

	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/eschercloudai/baski/pkg/trivy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type scanOptions struct {
	flags.GlobalFlags

	imageID           string
	flavorName        string
	networkID         string
	attachConfigDrive bool
	autoDeleteImage   bool
	skipCVECheck      bool
	maxSeverityScore  float64
	maxSeverityType   string
}

func (o *scanOptions) addFlags(cmd *cobra.Command) {
	viperPrefix := "scan"

	o.GlobalFlags.AddFlags(cmd)

	flags.StringVarWithViper(cmd, &o.flavorName, viperPrefix, "flavor-name", "", "The flavor of instance to build for scanning the image")
	flags.StringVarWithViper(cmd, &o.imageID, viperPrefix, "image-id", "", "The ID of the image to scan")
	flags.StringVarWithViper(cmd, &o.networkID, viperPrefix, "network-id", "", "Network ID to deploy the server onto for scanning")
	flags.BoolVarWithViper(cmd, &o.attachConfigDrive, viperPrefix, "attach-config-drive", false, "Used to enable a config drive on Openstack - this may be required if using an external network")
	flags.BoolVarWithViper(cmd, &o.autoDeleteImage, viperPrefix, "auto-delete-image", false, "If true, the image will be deleted if a vulnerability check does not succeed - recommended when building new images.")
	flags.BoolVarWithViper(cmd, &o.skipCVECheck, viperPrefix, "skip-cve-check", false, "If true, the image will be allowed even if a vulnerability is detected.")
	flags.Float64VarWithViper(cmd, &o.maxSeverityScore, viperPrefix, "max-severity-score", 7.0, "Can be anything from 0.1 to 10.0. Anything equal to or above this value will cause a failure. (Unless skip-cve-check is supplied)")
	flags.StringVarWithViper(cmd, &o.maxSeverityType, viperPrefix, "max-severity-type", "MEDIUM", "Accepted values are NONE, LOW, MEDIUM, HIGH, CRITICAL. This value will be what the score is checked against For example, a LOW 7.0 would be ignored if the value was HIGH with a `max-severity-score` of 7.0. (Unless skip-cve-check is supplied)")
}

// NewScanCommand creates a command that allows the scanning of an image.
func NewScanCommand() *cobra.Command {
	o := &scanOptions{}

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

			if !trivy.ValidSeverity(strings.ToUpper(viper.GetString("scan.max-severity-type"))) {
				log.Fatalln("severity value passed is invalid. Allowed values are: NONE, LOW, MEDIUM, HIGH, CRITICAL")
			}

			cloudsConfig := ostack.InitOpenstack()
			cloudsConfig.SetOpenstackEnvs()

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[viper.GetString("cloud-name")])

			kp := osClient.CreateKeypair(viper.GetString("scan.image-id"))
			server, freeIP := osClient.CreateServer(kp, viper.GetString("scan.image-id"), viper.GetString("scan.flavor-name"), viper.GetString("scan.network-id"), viper.GetBool("scan.attach-config-drive"))

			err := FetchResultsFromServer(freeIP, kp)
			if err != nil {
				RemoveScanningResources(server.ID, kp.Name, osClient)
				log.Fatalln(err.Error())
			}
			if !viper.GetBool("scan.skip-cve-check") {
				scoreCheck := CheckForVulnerabilities(viper.GetFloat64("scan.max-severity-score"), strings.ToUpper(viper.GetString("scan.max-severity-type")))
				if len(scoreCheck) != 0 {
					// Cleanup the scanning resources
					RemoveScanningResources(server.ID, kp.Name, osClient)

					if viper.GetBool("scan.auto-delete-image") {
						osClient.RemoveImage(viper.GetString("scan.image-id"))
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

					log.Fatalln("Vulnerabilities detected above threshold - removed image from infra. Please see the possible fixes located at '/tmp/results.json' for further information on this.")
				}
			}

			// Cleanup the scanning resources
			RemoveScanningResources(server.ID, kp.Name, osClient)
		},
	}

	o.addFlags(cmd)

	return cmd
}
