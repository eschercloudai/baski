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

package scan

import (
	"github.com/drew-viles/baskio/pkg/constants"
	ostack "github.com/drew-viles/baskio/pkg/openstack"
	"github.com/spf13/cobra"
	"log"
)

var (
	imageIDFlag               string
	networkIDFlag, flavorFlag string
	enableConfigDriveFlag     bool
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
			constants.Envs.SetOpenstackEnvs()

			osClient := &ostack.Client{
				Env: constants.Envs,
			}

			osClient.OpenstackInit()

			kp := osClient.CreateKeypair(imageIDFlag)
			server, freeIP := osClient.CreateServer(kp, imageIDFlag, flavorFlag, networkIDFlag, enableConfigDriveFlag)

			resultsFile, err := fetchResultsFromServer(freeIP, kp)
			if err != nil {
				removeScanningResources(server.ID, kp.Name, osClient)
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			//Cleanup the scanning resources
			removeScanningResources(server.ID, kp.Name, osClient)
		},
	}

	cmd.Flags().StringVarP(&flavorFlag, "instance-flavor", "f", "", "The flavor of instance to build for scanning the image.")
	cmd.Flags().StringVarP(&imageIDFlag, "image-id", "i", "", "The ID of the image to scan.")
	cmd.Flags().StringVarP(&networkIDFlag, "network-id", "n", "", "Network ID to deploy the server onto for scanning.")
	cmd.Flags().BoolVarP(&enableConfigDriveFlag, "enable-config-drive", "d", false, "Used to enable a config drive on Openstack. This may be required if using an external network.")

	requireFlag(cmd, "image-id")
	requireFlag(cmd, "instance-flavor")
	requireFlag(cmd, "network-id")

	return cmd
}

// requireFlag sets flags as required.
func requireFlag(cmd *cobra.Command, flag string) {
	err := cmd.MarkFlagRequired(flag)
	if err != nil {
		log.Fatalln(err)
	}
}
