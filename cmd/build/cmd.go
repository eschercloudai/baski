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

package build

import (
	"fmt"
	"github.com/drew-viles/baskio/pkg/constants"
	ostack "github.com/drew-viles/baskio/pkg/openstack"
	"github.com/spf13/cobra"
	"log"
	"path/filepath"
	"strings"
)

var (
	repoRoot                   = "https://github.com/eschercloudai/image-builder"
	imageRepoFlag, buildOSFlag string
	addGPUSupportFlag          bool
	gpuVersionFlag             = "510.73.08" // we'll flag this up later once the image builder supports it.

	networkIDFlag, openstackBuildConfigPathFlag string
	enableConfigDriveFlag                       bool
)

// NewBuildCommand creates a command that allows the building of an image.
func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build image",
		Long: `Build image.

Building images requires a set of commands to be run on the terminal however this is tedious and time consuming.
By using this, the time is reduced and automation can be enabled.

Overtime this will become more dynamic to allow for build customised 
images such as ones with GPU/HPC drivers/tools.

To use baskio to build an image, an Openstack cluster is required.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !checkValidOSSelected() {
				log.Fatalf("an unsupported OS has been entered. Please select a valid OS: %s\n", constants.SupportedOS)
			}

			constants.Envs.SetOpenstackEnvs()

			buildConfig := ostack.ParseBuildConfig(openstackBuildConfigPathFlag)
			buildConfig.ImageName = generateImageName(buildOSFlag, buildConfig.KubernetesSemver, addGPUSupportFlag, gpuVersionFlag)
			buildConfig.Networks = networkIDFlag

			buildGitDir := fetchBuildRepo(imageRepoFlag, addGPUSupportFlag)

			generateVariablesFile(buildGitDir, buildConfig)

			capiPath := filepath.Join(buildGitDir, "images/capi")
			fetchDependencies(capiPath)
			err := buildImage(capiPath, buildOSFlag)
			if err != nil {
				log.Fatalln(err)
			}

			imgID, err := retrieveNewImageID()
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println(imgID)
		},
	}

	cmd.Flags().StringVarP(&openstackBuildConfigPathFlag, "build-config", "c", "", strings.Join([]string{"The openstack packer variables file to use to build the image"}, ""))
	cmd.Flags().StringVarP(&buildOSFlag, "build-os", "o", "ubuntu-2204", "This is the target os to build. Valid values are currently: ubuntu-2004 and ubuntu-2204")
	cmd.Flags().BoolVarP(&enableConfigDriveFlag, "enable-config-drive", "d", false, "Used to enable a config drive on Openstack. This may be required if using an external network.")
	cmd.Flags().BoolVarP(&addGPUSupportFlag, "enable-gpu-support", "g", false, "This will configure GPU support in the image")
	cmd.Flags().StringVarP(&imageRepoFlag, "imageRepo", "r", strings.Join([]string{repoRoot, "git"}, "."), "The imageRepo from which the image builder should be deployed.")
	cmd.Flags().StringVarP(&networkIDFlag, "network-id", "n", "", "Network ID to deploy the server onto for scanning.")

	requireFlag(cmd, "build-config")
	requireFlag(cmd, "network-id")

	return cmd
}

// checkValidOSSelected checks that the build os provided is a valid one.
func checkValidOSSelected() bool {
	for _, v := range constants.SupportedOS {
		if buildOSFlag == v {
			return true
		}
	}
	return false
}

// requireFlag sets flags as required.
func requireFlag(cmd *cobra.Command, flag string) {
	err := cmd.MarkFlagRequired(flag)
	if err != nil {
		log.Fatalln(err)
	}
}
