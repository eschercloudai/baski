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

package flags

import (
	"fmt"
	"github.com/eschercloudai/baski/pkg/util/completion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// OpenStackCoreFlags are the core requirements for any interaction with the openstack cloud.
type OpenStackCoreFlags struct {
	CloudsPath string
	CloudName  string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct.
func (o *OpenStackCoreFlags) SetOptionsFromViper() {
	o.CloudsPath = viper.GetString(fmt.Sprintf("%s.clouds-file", viperOpenStackPrefix))
	o.CloudName = viper.GetString(fmt.Sprintf("%s.cloud-name", viperOpenStackPrefix))
}

func (o *OpenStackCoreFlags) AddFlags(cmd *cobra.Command, viperPrefix string) {
	PersistentStringVarWithViper(cmd, &o.CloudsPath, viperPrefix, "clouds-file", "~/.config/openstack/clouds.yaml", "The location of the openstack clouds.yaml file to use")
	PersistentStringVarWithViper(cmd, &o.CloudName, viperPrefix, "cloud-name", "", "The name of the cloud profile to use from the clouds.yaml file")
	if err := cmd.RegisterFlagCompletionFunc("cloud-name", completion.CloudCompletionFunc); err != nil {
		panic(err)
	}
	cmd.MarkFlagsRequiredTogether("clouds-file", "cloud-name")
}

// OpenStackInstanceFlags are Additional flags that can would be required for other steps such as scan, sign and publish.
type OpenStackInstanceFlags struct {
	AttachConfigDrive bool
	NetworkID         string
	FlavorName        string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct.
func (o *OpenStackInstanceFlags) SetOptionsFromViper() {
	o.NetworkID = viper.GetString(fmt.Sprintf("%s.network-id", viperOpenStackPrefix))
	o.FlavorName = viper.GetString(fmt.Sprintf("%s.flavor-name", viperOpenStackPrefix))
	o.AttachConfigDrive = viper.GetBool(fmt.Sprintf("%s.attach-config-drive", viperOpenStackPrefix))
}

func (o *OpenStackInstanceFlags) AddFlags(cmd *cobra.Command, viperPrefix string) {
	StringVarWithViper(cmd, &o.NetworkID, viperPrefix, "network-id", "", "Network ID to deploy the server onto for scanning")
	StringVarWithViper(cmd, &o.FlavorName, viperPrefix, "flavor-name", "", "The Name of the instance flavor to use for the build")
	BoolVarWithViper(cmd, &o.AttachConfigDrive, viperPrefix, "attach-config-drive", false, "Used to enable a config drive on Openstack. This may be required if directly attaching an external network to the instance")
}

// OpenStackFlags are explicitly for OpenStack based clouds and defines settings that pertain only to that cloud type.
type OpenStackFlags struct {
	OpenStackCoreFlags
	OpenStackInstanceFlags

	SourceImageID         string
	UseFloatingIP         bool
	FloatingIPNetworkName string
	ImageVisibility       string
	ImageDiskFormat       string
	VolumeType            string
	VolumeSize            int
	RootfsUUID            string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct.
func (o *OpenStackFlags) SetOptionsFromViper() {
	o.OpenStackCoreFlags.SetOptionsFromViper()
	o.OpenStackInstanceFlags.SetOptionsFromViper()

	o.SourceImageID = viper.GetString(fmt.Sprintf("%s.source-image", viperOpenStackPrefix))
	o.UseFloatingIP = viper.GetBool(fmt.Sprintf("%s.use-floating-ip", viperOpenStackPrefix))
	o.FloatingIPNetworkName = viper.GetString(fmt.Sprintf("%s.floating-ip-network-name", viperOpenStackPrefix))
	o.ImageVisibility = viper.GetString(fmt.Sprintf("%s.image-visibility", viperOpenStackPrefix))
	o.ImageDiskFormat = viper.GetString(fmt.Sprintf("%s.image-disk-format", viperOpenStackPrefix))
	o.VolumeType = viper.GetString(fmt.Sprintf("%s.volume-type", viperOpenStackPrefix))
	o.VolumeSize = viper.GetInt(fmt.Sprintf("%s.volume-size", viperOpenStackPrefix))
	o.RootfsUUID = viper.GetString(fmt.Sprintf("%s.rootfs-uuid", viperOpenStackPrefix))
}

func (o *OpenStackFlags) AddFlags(cmd *cobra.Command, viperPrefix string) {
	o.OpenStackCoreFlags.AddFlags(cmd, viperPrefix)
	o.OpenStackInstanceFlags.AddFlags(cmd, viperPrefix)

	StringVarWithViper(cmd, &o.SourceImageID, viperPrefix, "source-image-id", "ubuntu-2204", "The ID of the image that will be used as a base for the newly built image")
	BoolVarWithViper(cmd, &o.UseFloatingIP, viperPrefix, "use-floating-ip", true, "Whether to use a floating IP for the instance")
	StringVarWithViper(cmd, &o.FloatingIPNetworkName, viperPrefix, "floating-ip-network-name", "Internet", "The Name of the network in which to create the floating ip")
	StringVarWithViper(cmd, &o.ImageVisibility, viperPrefix, "image-visibility", "private", "Change the image visibility in Openstack - you need to ensure the use you're authenticating with has permissions to do so or this will fail")
	StringVarWithViper(cmd, &o.ImageDiskFormat, viperPrefix, "image-disk-format", "raw", "The image disk format in Openstack")
	StringVarWithViper(cmd, &o.VolumeType, viperPrefix, "volume-type", "", "The volume type in Openstack")
	IntVarWithViper(cmd, &o.VolumeSize, viperPrefix, "volume-size", 10, "Size of the Block Storage service volume in GB")
	StringVarWithViper(cmd, &o.RootfsUUID, viperPrefix, "rootfs-uuid", "", "The UUID of the root filesystem. This can be acquired from the source image that the resulting image will be built from - (this will be automated soon™)")

	cmd.MarkFlagsRequiredTogether("use-floating-ip", "floating-ip-network-name")
}
