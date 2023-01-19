/*
Copyright 2022 EscherCloud.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/eschercloudai/baskio/cmd/scan"
	ostack "github.com/eschercloudai/baskio/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

// NewScanCommand creates a command that allows the scanning of an image.
func NewScanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan image",
		Long: `Scan image.

Scanning an image requires the creation of a new instance in Openstack using the image you want to scan.
Then, Trivy needs downloading and running against the filesystem. Again, this is time consuming.

The scan section of baskio fixes this for you and allows you to drink coffee whilst it does the hard work for you.

It creates a new instance using the provided Openstack configuration variables and scans the image.
Once complete, it generates a report file that you can read,
OR!
Use the publish command to create a "pretty" interface in GitHub Pages through which you can browse the results.`,
		Run: func(cmd *cobra.Command, args []string) {
			cloudsConfig := ostack.InitOpenstack()
			cloudsConfig.SetOpenstackEnvs()

			osClient := &ostack.Client{
				Cloud: cloudsConfig.Clouds[viper.GetString("cloud-name")],
			}
			osClient.OpenstackInit()

			kp := osClient.CreateKeypair(viper.GetString("scan.image-id"))
			server, freeIP := osClient.CreateServer(kp, viper.GetString("scan.image-id"), viper.GetString("scan.flavor-name"), viper.GetString("scan.network-id"), viper.GetBool("scan.attach-config-drive"))

			resultsFile, err := scan.FetchResultsFromServer(freeIP, kp)
			if err != nil {
				scan.RemoveScanningResources(server.ID, kp.Name, osClient)
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			//Cleanup the scanning resources
			scan.RemoveScanningResources(server.ID, kp.Name, osClient)
		},
	}

	cmd.Flags().StringVar(&flavorNameFlag, "flavor-name", "", "The flavor of instance to build for scanning the image")
	cmd.Flags().StringVar(&imageIDFlag, "image-id", "", "The ID of the image to scan")
	cmd.Flags().StringVar(&networkIDFlag, "network-id", "", "Network ID to deploy the server onto for scanning")
	cmd.Flags().BoolVar(&attachConfigDriveFlag, "attach-config-drive", false, "Used to enable a config drive on Openstack - this may be required if using an external network")

	bindViper(cmd, "scan.flavor-name", "flavor-name")
	bindViper(cmd, "scan.image-id", "image-id")
	bindViper(cmd, "scan.network-id", "network-id")
	bindViper(cmd, "scan.attach-config-drive", "attach-config-drive")

	return cmd
}
