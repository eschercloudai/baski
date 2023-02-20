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
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

type scanOptions struct {
	flags.GlobalFlags

	imageID           string
	flavorName        string
	networkID         string
	attachConfigDrive bool
}

func (o *scanOptions) addFlags(cmd *cobra.Command) {
	viperPrefix := "scan"

	o.GlobalFlags.AddFlags(cmd)

	flags.StringVarWithViper(cmd, &o.flavorName, viperPrefix, "flavor-name", "", "The flavor of instance to build for scanning the image")
	flags.StringVarWithViper(cmd, &o.imageID, viperPrefix, "image-id", "", "The ID of the image to scan")
	flags.StringVarWithViper(cmd, &o.networkID, viperPrefix, "network-id", "", "Network ID to deploy the server onto for scanning")
	flags.BoolVarWithViper(cmd, &o.attachConfigDrive, viperPrefix, "attach-config-drive", false, "Used to enable a config drive on Openstack - this may be required if using an external network")
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

The scan section of baski fixes this for you and allows you to drink coffee whilst it does the hard work for you.

It creates a new instance using the provided Openstack configuration variables and scans the image.
Once complete, it generates a report file that you can read,
OR!
Use the publish command to create a "pretty" interface in GitHub Pages through which you can browse the results.`,
		Run: func(cmd *cobra.Command, args []string) {
			cloudsConfig := ostack.InitOpenstack()
			cloudsConfig.SetOpenstackEnvs()

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[viper.GetString("cloud-name")])

			kp := osClient.CreateKeypair(viper.GetString("scan.image-id"))
			server, freeIP := osClient.CreateServer(kp, viper.GetString("scan.image-id"), viper.GetString("scan.flavor-name"), viper.GetString("scan.network-id"), viper.GetBool("scan.attach-config-drive"))

			resultsFile, err := FetchResultsFromServer(freeIP, kp)
			if err != nil {
				RemoveScanningResources(server.ID, kp.Name, osClient)
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			//Cleanup the scanning resources
			RemoveScanningResources(server.ID, kp.Name, osClient)
		},
	}

	o.addFlags(cmd)

	return cmd
}
