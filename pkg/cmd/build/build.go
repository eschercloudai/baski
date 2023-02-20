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

package build

import (
	"github.com/eschercloudai/baski/pkg/cmd/util/data"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/eschercloudai/baski/pkg/constants"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"strings"
)

var (
	repoRoot = "https://github.com/eschercloudai/image-builder"
)

type buildOptions struct {
	flags.GlobalFlags

	verbose               bool
	buildOS               string
	imageRepo             string
	networkID             string
	sourceImageID         string
	flavorName            string
	userFloatingIP        bool
	floatingIPNetworkName string
	attachConfigDrive     bool
	imageVisibility       string
	crictlVersion         string
	cniVersion            string
	kubeVersion           string
	extraDebs             string
	addNvidiaSupport      bool
	nvidiaVersion         string
	nvidiaInstallerURL    string
	nvidiaTOKURL          string
	griddFeatureType      int
	rootfsUUID            string
}

func (o *buildOptions) addFlags(cmd *cobra.Command) {
	viperPrefix := "build"

	o.GlobalFlags.AddFlags(cmd)
	flags.StringVarWithViper(cmd, &o.buildOS, viperPrefix, "build-os", "ubuntu-2204", "This is the target os to build. Valid values are currently: ubuntu-2004 and ubuntu-2204")
	flags.BoolVarWithViper(cmd, &o.attachConfigDrive, viperPrefix, "attach-config-drive", false, "Used to enable a config drive on Openstack. This may be required if directly attaching an external network to the instance")
	flags.StringVarWithViper(cmd, &o.imageRepo, viperPrefix, "image-repo", strings.Join([]string{repoRoot, "git"}, "."), "The imageRepo from which the image builder should be deployed")
	flags.StringVarWithViper(cmd, &o.sourceImageID, viperPrefix, "source-image-id", "ubuntu-2204", "The ID of the image that will be used as a base for the newly built image")
	flags.StringVarWithViper(cmd, &o.networkID, viperPrefix, "network-id", "", "Network ID to deploy the server onto for scanning")
	flags.StringVarWithViper(cmd, &o.flavorName, viperPrefix, "flavor-name", "", "The Name of the instance flavor to use for the build")
	flags.BoolVarWithViper(cmd, &o.userFloatingIP, viperPrefix, "use-floating-ip", true, "Whether to use a floating IP for the instance")
	flags.StringVarWithViper(cmd, &o.floatingIPNetworkName, viperPrefix, "floating-ip-network-name", "Internet", "The Name of the network in which to create the floating ip")
	flags.StringVarWithViper(cmd, &o.imageVisibility, viperPrefix, "image-visibility", "private", "Change the image visibility in Openstack - you need to ensure the use you're authenticating with has permissions to do so or this will fail")
	flags.StringVarWithViper(cmd, &o.cniVersion, viperPrefix, "cni-version", "1.1.1", "The CNI plugins version to include to the built image")
	flags.StringVarWithViper(cmd, &o.crictlVersion, viperPrefix, "crictl-version", "1.25.0", "The crictl-tools version to add to the built image")
	flags.StringVarWithViper(cmd, &o.kubeVersion, viperPrefix, "kubernetes-version", "1.25.3", "The Kubernetes version to add to the built image")
	flags.StringVarWithViper(cmd, &o.extraDebs, viperPrefix, "extra-debs", "", "A space-seperated list of any extra (Debian / Ubuntu) packages that should be installed")
	flags.StringVarWithViper(cmd, &o.rootfsUUID, viperPrefix, "rootfs-uuid", "", "The UUID of the root filesystem. This can be acquired from the source image that the resulting image will be built from - (this will be automated soon™)")
	flags.BoolVarWithViper(cmd, &o.addNvidiaSupport, viperPrefix, "enable-nvidia-support", false, "This will configure Nvidia support in the image")
	flags.StringVarWithViper(cmd, &o.nvidiaVersion, viperPrefix, "nvidia-driver-version", "510.73.08", "The Nvidia driver version")
	flags.StringVarWithViper(cmd, &o.nvidiaInstallerURL, viperPrefix, "nvidia-installer-url", "", "The Nvidia installer download URL - this must be acquired from Nvidia")
	flags.StringVarWithViper(cmd, &o.nvidiaTOKURL, viperPrefix, "nvidia-tok-url", "", "The Nvidia .tok file download URL - this must be acquired from Nvidia")
	flags.IntVarWithViper(cmd, &o.griddFeatureType, viperPrefix, "gridd-feature-type", 4, "The gridd feature type - See https://docs.nvidia.com/license-system/latest/nvidia-license-system-quick-start-guide/index.html#configuring-nls-licensed-client-on-linux for more information")
	flags.BoolVarWithViper(cmd, &o.verbose, viperPrefix, "verbose", false, "Enable verbose output to see the information from packer. Not turning this on will mean the process appears to hang while the image build happens.")

	cmd.MarkFlagsRequiredTogether("enable-nvidia-support", "nvidia-tok-url", "nvidia-installer-url", "nvidia-tok-url")
	cmd.MarkFlagsRequiredTogether("use-floating-ip", "floating-ip-network-name")
	cmd.MarkFlagsRequiredTogether("crictl-version", "kubernetes-version")
}

// NewBuildCommand creates a command that allows the building of an image.
func NewBuildCommand() *cobra.Command {
	o := &buildOptions{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build image",
		Long: `Build image.

Building images requires a set of commands to be run on the terminal however this is tedious and time consuming.
By using this, the time is reduced and automation can be enabled.

Overtime this will become more dynamic to allow for build customised 
images such as ones with GPU/HPC drivers/tools.

To use baski to build an image, an Openstack cluster is required.`,
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			cloudsConfig := ostack.InitOpenstack()
			packerBuildConfig := ostack.InitPackerConfig()
			if !checkValidOSSelected(o.buildOS) {
				log.Fatalf("an unsupported OS has been entered. Please select a valid OS: %s\n", constants.SupportedOS)
			}

			buildGitDir := CreateRepoDirectory()
			FetchBuildRepo(buildGitDir, viper.GetString("build.image-repo"), viper.GetBool("build.enable-nvidia-support"))

			metadata := ostack.GenerateBuilderMetadata()
			ostack.UpdatePackerBuildersJson(buildGitDir, metadata)

			capiPath := filepath.Join(buildGitDir, "images", "capi")
			packerBuildConfig.GenerateVariablesFile(capiPath)

			InstallDependencies(capiPath)

			cloudsConfig.SetOpenstackEnvs()

			err := BuildImage(capiPath, viper.GetString("build.build-os"))
			if err != nil {
				log.Fatalln(err)
			}

			imgID, err := data.RetrieveNewImageID()
			if err != nil {
				log.Fatalln(err)
			}

			err = SaveImageIDToFile(imgID)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	o.addFlags(cmd)

	return cmd
}

// checkValidOSSelected checks that the build os provided is a valid one.
func checkValidOSSelected(buildOS string) bool {
	for _, v := range constants.SupportedOS {
		if buildOS == v {
			return true
		}
	}
	return false
}
