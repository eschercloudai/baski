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
	"log"
	"path/filepath"
	"strings"

	"github.com/drew-viles/baskio/cmd/build"
	"github.com/drew-viles/baskio/pkg/constants"
	ostack "github.com/drew-viles/baskio/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	repoRoot = "https://github.com/eschercloudai/image-builder"
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
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			cloudsConfig := ostack.InitOpenstack()
			packerBuildConfig := ostack.InitPackerConfig()
			if !checkValidOSSelected() {
				log.Fatalf("an unsupported OS has been entered. Please select a valid OS: %s\n", constants.SupportedOS)
			}

			buildGitDir := build.CreateRepoDirectory()
			build.FetchBuildRepo(buildGitDir, imageRepoFlag, viper.GetBool("build.enable-nvidia-support"))

			capiPath := filepath.Join(buildGitDir, "images", "capi")
			packerBuildConfig.GenerateVariablesFile(capiPath)

			build.InstallDependencies(capiPath)

			cloudsConfig.SetOpenstackEnvs()

			err := build.BuildImage(capiPath, viper.GetString("build.build-os"))
			if err != nil {
				log.Fatalln(err)
			}

			imgID, err := build.RetrieveNewImageID()
			if err != nil {
				log.Fatalln(err)
			}

			err = build.SaveImageIDToFile(imgID)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringVar(&buildOSFlag, "build-os", "ubuntu-2204", "This is the target os to build. Valid values are currently: ubuntu-2004 and ubuntu-2204")
	cmd.Flags().BoolVar(&attachConfigDriveFlag, "attach-config-drive", false, "Used to enable a config drive on Openstack. This may be required if directly attaching an external network to the instance")
	cmd.Flags().StringVar(&imageRepoFlag, "image-repo", strings.Join([]string{repoRoot, "git"}, "."), "The imageRepo from which the image builder should be deployed")
	cmd.Flags().StringVar(&sourceImageIDFlag, "source-image-id", "ubuntu-2204", "The ID of the image that will be used as a base for the newly built image")
	cmd.Flags().StringVar(&networkIDFlag, "network-id", "", "Network ID to deploy the server onto for scanning")
	cmd.Flags().StringVar(&flavorNameFlag, "flavor-name", "", "The Name of the instance flavor to use for the build")
	cmd.Flags().BoolVar(&userFloatingIPFlag, "use-floating-ip", true, "Whether to use a floating IP for the instance")
	cmd.Flags().StringVar(&floatingIPNetworkNameFlag, "floating-ip-network-name", "Internet", "The Name of the network in which to create the floating ip")
	cmd.Flags().StringVar(&imageVisibilityFlag, "image-visibility", "private", "Change the image visibility in Openstack - you need to ensure the use you're authenticating with has permissions to do so or this will fail")
	cmd.Flags().StringVar(&crictlVersionFlag, "crictl-version", "1.25.0", "The crictl-tools version to add to the built image")
	cmd.Flags().StringVar(&kubeVersionFlag, "kubernetes-version", "1.25.3", "The Kubernetes version to add to the built image")
	cmd.Flags().StringVar(&extraDebsFlag, "extra-debs", "", "A comma-seperated list of any extra (Debian / Ubuntu) packages that should be installed")
	cmd.Flags().BoolVar(&addNvidiaSupportFlag, "enable-nvidia-support", false, "This will configure Nvidia support in the image")
	cmd.Flags().StringVar(&nvidiaInstallerURLFlag, "nvidia-installer-url", "", "The Nvidia installer location - this must be acquired from Nvidia")
	cmd.Flags().StringVar(&nvidiaVersionFlag, "nvidia-driver-version", "510.73.08", "The Nvidia driver version")
	cmd.Flags().StringVar(&gridLicenseServerFlag, "grid-license-server", "", "The url or address of the licensing server to pull the gridd.conf from")
	cmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Enable verbose output to see the information from packer. Not turning this on will mean the process appears to hang while the image build happens.")

	cmd.MarkFlagsRequiredTogether("enable-nvidia-support", "grid-license-server", "nvidia-installer-url")
	cmd.MarkFlagsRequiredTogether("use-floating-ip", "floating-ip-network-name")
	cmd.MarkFlagsRequiredTogether("crictl-version", "kubernetes-version")

	bindViper(cmd, "build.verbose", "verbose")
	bindViper(cmd, "build.build-os", "build-os")
	bindViper(cmd, "build.attach-config-drive", "attach-config-drive")
	bindViper(cmd, "build.image-repo", "image-repo")
	bindViper(cmd, "build.source-image-id", "source-image-id")
	bindViper(cmd, "build.network-id", "network-id")
	bindViper(cmd, "build.flavor-name", "flavor-name")
	bindViper(cmd, "build.use-floating-ip", "use-floating-ip")
	bindViper(cmd, "build.floating-ip-network-name", "floating-ip-network-name")
	bindViper(cmd, "build.image-visibility", "image-visibility")
	bindViper(cmd, "build.crictl-version", "crictl-version")
	bindViper(cmd, "build.kubernetes-version", "kubernetes-version")
	bindViper(cmd, "build.extra-debs", "extra-debs")
	bindViper(cmd, "build.enable-nvidia-support", "enable-nvidia-support")
	bindViper(cmd, "build.grid-license-server", "grid-license-server")
	bindViper(cmd, "build.nvidia-installer-url", "nvidia-installer-url")
	bindViper(cmd, "build.nvidia-driver-version", "nvidia-driver-version")

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
