/*
Copyright 2024 Drewbernetes.

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
	"github.com/drewbernetes/baski/pkg/util/completion"
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
	BoolVarWithViper(cmd, &o.AttachConfigDrive, viperPrefix, "attach-config-drive", false, "Whether or not nova should use ConfigDrive for cloud-init metadata.")
}

// OpenStackFlags are explicitly for OpenStack based clouds and defines settings that pertain only to that cloud type.
type OpenStackFlags struct {
	OpenStackCoreFlags
	OpenStackInstanceFlags

	SourceImageID         string
	SSHPrivateKeyFile     string
	SSHKeypairName        string
	UseFloatingIP         bool
	FloatingIPNetworkName string
	SecurityGroup         string
	ImageVisibility       string
	ImageDiskFormat       string
	UseBlockStorageVolume string
	VolumeType            string
	VolumeSize            int
	RootfsUUID            string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct.
func (q *OpenStackFlags) SetOptionsFromViper() {
	q.SourceImageID = viper.GetString(fmt.Sprintf("%s.source-image", viperOpenStackPrefix))
	q.SSHPrivateKeyFile = viper.GetString(fmt.Sprintf("%s.ssh-privatekey-file", viperOpenStackPrefix))
	q.SSHKeypairName = viper.GetString(fmt.Sprintf("%s.ssh-keypair-name", viperOpenStackPrefix))
	q.UseFloatingIP = viper.GetBool(fmt.Sprintf("%s.use-floating-ip", viperOpenStackPrefix))
	q.FloatingIPNetworkName = viper.GetString(fmt.Sprintf("%s.floating-ip-network-name", viperOpenStackPrefix))
	q.SecurityGroup = viper.GetString(fmt.Sprintf("%s.security-group", viperOpenStackPrefix))
	q.ImageVisibility = viper.GetString(fmt.Sprintf("%s.image-visibility", viperOpenStackPrefix))
	q.ImageDiskFormat = viper.GetString(fmt.Sprintf("%s.image-disk-format", viperOpenStackPrefix))
	q.UseBlockStorageVolume = viper.GetString(fmt.Sprintf("%s.use-blockstorage-volume", viperOpenStackPrefix))
	q.VolumeType = viper.GetString(fmt.Sprintf("%s.volume-type", viperOpenStackPrefix))
	q.VolumeSize = viper.GetInt(fmt.Sprintf("%s.volume-size", viperOpenStackPrefix))
	q.RootfsUUID = viper.GetString(fmt.Sprintf("%s.rootfs-uuid", viperOpenStackPrefix))

	q.OpenStackCoreFlags.SetOptionsFromViper()
	q.OpenStackInstanceFlags.SetOptionsFromViper()
}

func (q *OpenStackFlags) AddFlags(cmd *cobra.Command, viperPrefix string) {
	StringVarWithViper(cmd, &q.SourceImageID, viperPrefix, "source-image-id", "ubuntu-2204", "The ID of the image that will be used as a base for the newly built image")
	StringVarWithViper(cmd, &q.SSHPrivateKeyFile, viperPrefix, "ssh-privatekey-file", "", "The Private Key to use when using ssh-keypair-name")
	StringVarWithViper(cmd, &q.SSHKeypairName, viperPrefix, "ssh-keypair-name", "", "Define an SSH Keypair to use instead of generating automatically")
	BoolVarWithViper(cmd, &q.UseFloatingIP, viperPrefix, "use-floating-ip", true, "Whether to use a floating IP for the instance")
	StringVarWithViper(cmd, &q.FloatingIPNetworkName, viperPrefix, "floating-ip-network-name", "public1", "The Name of the network in which to create the floating ip")
	StringVarWithViper(cmd, &q.SecurityGroup, viperPrefix, "security-group", "", "Specify the security groups to attach")
	StringVarWithViper(cmd, &q.ImageVisibility, viperPrefix, "image-visibility", "private", "Change the image visibility in Openstack - you need to ensure the use you're authenticating with has permissions to do so or this will fail")
	StringVarWithViper(cmd, &q.ImageDiskFormat, viperPrefix, "image-disk-format", "", "The image disk format in Openstack")
	StringVarWithViper(cmd, &q.UseBlockStorageVolume, viperPrefix, "use-blockstorage-volume", "", "Use Block Storage service volume for the instance root volume instead of Compute service local volume")
	StringVarWithViper(cmd, &q.VolumeType, viperPrefix, "volume-type", "", "Type of the Block Storage service volume. If this isn't specified, the default enforced by your OpenStack cluster will be used")
	IntVarWithViper(cmd, &q.VolumeSize, viperPrefix, "volume-size", 0, "Size of the Block Storage service volume in GB. If this isn't specified, it is set to source image min disk value (if set) or calculated from the source image bytes size. Note that in some cases this needs to be specified, if use_blockstorage_volume is true")
	StringVarWithViper(cmd, &q.RootfsUUID, viperPrefix, "rootfs-uuid", "", "The UUID of the root filesystem. This can be acquired from the source image that the resulting image will be built from - (this will be automated soon™)")

	q.OpenStackCoreFlags.AddFlags(cmd, viperPrefix)
	q.OpenStackInstanceFlags.AddFlags(cmd, viperPrefix)

	cmd.MarkFlagsRequiredTogether("use-floating-ip", "floating-ip-network-name")
}
