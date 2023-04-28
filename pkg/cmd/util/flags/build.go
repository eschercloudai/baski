package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

type BuildOptions struct {
	OpenStackFlags
	//Add more clouds here

	// Verbose will output all output from the make command if set to true.
	Verbose bool
	// BuildOS is used to denote which Operating system to use. See the image builder for valid values for the cloud being used.
	BuildOS string
	// ImagePrefix is a string that is prepended onto the name of the image.
	ImagePrefix string
	// ImageRepo is used to override the repo used for the image build. It defaults to the kubernetes-sigs/image-builder repo.
	ImageRepo string
	// CrictlVersion denotes the version of cri-tools to install.
	CrictlVersion string
	// CNIVersion denotes the CNI version to install.
	CniVersion string
	// KubeVersion denotes the version of Kubernetes to install.
	KubeVersion string
	// ExtraDebs enables the installation of extra packages to be installed via the package manager - currently apt only.
	ExtraDebs string
	// AddFalco installs Falco onto the target image. This enables security features provided by Falco.
	AddFalco bool
	// AddTrivy installs Trivy onto the target image. This enables scanning to be performed using Trivy.
	AddTrivy bool
	// AddNvidiaSupport enables the installation of the NVidia driver - this must be used alongside all other NVida options as the driver is not publically available.
	AddNvidiaSupport bool
	// NvidiaVersion the version of NVidia being installed. This may be Deprecated soon as it's just used for tagging the image with metadata and could be pulled from the file name of the installer.
	NvidiaVersion string
	// NvidiaBucketEndpoint is used to define the S3 endpoint from which the Nvidia installer and TOK file would be downloaded.
	NvidiaBucketEndpoint string
	// NvidiaBucketName is the name of the bucket in the S3 storage from which the Nvidia installer and TOK file would be downloaded.
	NvidiaBucketName string
	// NvidiaBucketAccessKey contains the access key for accessing the S3 storage
	NvidiaBucketAccessKey string
	// NvidiaBucketSecretKey contains the secret key for accessing the S3 storage
	NvidiaBucketSecretKey string
	// NvidiaInstallerLocation contains the location in the bucket from which the .run file can be downloaded
	NvidiaInstallerLocation string
	// NvidiaTOKLocation contains the location in the bucket from which the .tok file can be downloaded.
	NvidiaTOKLocation string
	// NvidiaGriddFeatureType denotes the GRIDD FeatureType - https://docs.nvidia.com/grid/13.0/grid-licensing-user-guide/index.html#configuring-nls-licensed-client-on-linux
	NvidiaGriddFeatureType int
}

func (o *BuildOptions) SetOptionsFromViper() {
	o.OpenStackFlags.SetSignOptionsFromViper()

	// General Flags
	o.Verbose = viper.GetBool(fmt.Sprintf("%s.verbose", viperBuildPrefix))
	o.BuildOS = viper.GetString(fmt.Sprintf("%s.build-os", viperBuildPrefix))
	o.ImagePrefix = viper.GetString(fmt.Sprintf("%s.image-prefix", viperBuildPrefix))
	o.ImageRepo = viper.GetString(fmt.Sprintf("%s.image-repo", viperBuildPrefix))
	o.CrictlVersion = viper.GetString(fmt.Sprintf("%s.crictl-version", viperBuildPrefix))
	o.CniVersion = viper.GetString(fmt.Sprintf("%s.cni-version", viperBuildPrefix))
	o.KubeVersion = viper.GetString(fmt.Sprintf("%s.kubernetes-version", viperBuildPrefix))
	o.ExtraDebs = viper.GetString(fmt.Sprintf("%s.extra-debs", viperBuildPrefix))
	o.AddFalco = viper.GetBool(fmt.Sprintf("%s.add-falco", viperBuildPrefix))
	o.AddTrivy = viper.GetBool(fmt.Sprintf("%s.add-trivy", viperBuildPrefix))

	// NVIDIA
	o.AddNvidiaSupport = viper.GetBool(fmt.Sprintf("%s.enable-nvidia-support", viperNvidiaPrefix))
	o.NvidiaVersion = viper.GetString(fmt.Sprintf("%s.nvidia-driver-version", viperNvidiaPrefix))
	o.NvidiaBucketEndpoint = viper.GetString(fmt.Sprintf("%s.nvidia-bucket-endpoint", viperNvidiaPrefix))
	o.NvidiaBucketName = viper.GetString(fmt.Sprintf("%s.nvidia-bucket-name", viperNvidiaPrefix))
	o.NvidiaBucketAccessKey = viper.GetString(fmt.Sprintf("%s.nvidia-bucket-access", viperNvidiaPrefix))
	o.NvidiaBucketSecretKey = viper.GetString(fmt.Sprintf("%s.nvidia-bucket-secret", viperNvidiaPrefix))
	o.NvidiaInstallerLocation = viper.GetString(fmt.Sprintf("%s.nvidia-installer-location", viperNvidiaPrefix))
	o.NvidiaTOKLocation = viper.GetString(fmt.Sprintf("%s.nvidia-tok-location", viperNvidiaPrefix))
	o.NvidiaGriddFeatureType = viper.GetInt(fmt.Sprintf("%s.nvidia-gridd-feature-type", viperNvidiaPrefix))

}

func (o *BuildOptions) AddFlags(cmd *cobra.Command, imageBuilderRepo string) {
	o.OpenStackFlags.AddFlags(cmd, viperOpenStackPrefix)
	// Build flags
	StringVarWithViper(cmd, &o.BuildOS, viperBuildPrefix, "build-os", "ubuntu-2204", "This is the target os to build. Valid values are currently: ubuntu-2004 and ubuntu-2204")
	StringVarWithViper(cmd, &o.ImagePrefix, viperBuildPrefix, "image-prefix", "kube", "This will prefix the image with the value provided. Defaults to 'kube' producing an image name of kube-yymmdd-xxxxxxxx")
	StringVarWithViper(cmd, &o.ImageRepo, viperBuildPrefix, "image-repo", strings.Join([]string{imageBuilderRepo, "git"}, "."), "The imageRepo from which the image builder should be deployed")
	StringVarWithViper(cmd, &o.CniVersion, viperBuildPrefix, "cni-version", "1.2.0", "The CNI plugins version to include to the built image")
	StringVarWithViper(cmd, &o.CrictlVersion, viperBuildPrefix, "crictl-version", "1.25.0", "The crictl-tools version to add to the built image")
	StringVarWithViper(cmd, &o.KubeVersion, viperBuildPrefix, "kubernetes-version", "1.25.3", "The Kubernetes version to add to the built image")
	StringVarWithViper(cmd, &o.ExtraDebs, viperBuildPrefix, "extra-debs", "", "A space-seperated list of any extra (Debian / Ubuntu) packages that should be installed")
	BoolVarWithViper(cmd, &o.AddFalco, viperBuildPrefix, "add-falco", false, "If enabled, will install Falco onto the image")
	BoolVarWithViper(cmd, &o.AddTrivy, viperBuildPrefix, "add-trivy", false, "If enabled, will install Trivy onto the image")
	BoolVarWithViper(cmd, &o.Verbose, viperBuildPrefix, "verbose", false, "Enable verbose output to see the information from packer. Not turning this on will mean the process appears to hang while the image build happens")

	// NVIDIA flags
	BoolVarWithViper(cmd, &o.AddNvidiaSupport, viperNvidiaPrefix, "enable-nvidia-support", false, "This will configure NVIDIA support in the image")
	StringVarWithViper(cmd, &o.NvidiaVersion, viperNvidiaPrefix, "nvidia-driver-version", "525.85.05", "The NVIDIA driver version")
	StringVarWithViper(cmd, &o.NvidiaBucketEndpoint, viperNvidiaPrefix, "nvidia-bucket-endpoint", "", "The endpoint of the bucket from which to download the NVIDIA components")
	StringVarWithViper(cmd, &o.NvidiaBucketName, viperNvidiaPrefix, "nvidia-bucket-name", "", "The bucket name that the NVIDIA components are uploaded to")
	StringVarWithViper(cmd, &o.NvidiaBucketAccessKey, viperNvidiaPrefix, "nvidia-bucket-access", "", "The access key used to access the bucket from which to download the NVIDIA components")
	StringVarWithViper(cmd, &o.NvidiaBucketSecretKey, viperNvidiaPrefix, "nvidia-bucket-secret", "", "The secret key used to access the bucket from which to download the NVIDIA components")
	StringVarWithViper(cmd, &o.NvidiaInstallerLocation, viperNvidiaPrefix, "nvidia-installer-location", "", "The NVIDIA installer location in the bucket - this must be acquired from NVIDIA and uploaded to your bucket")
	StringVarWithViper(cmd, &o.NvidiaTOKLocation, viperNvidiaPrefix, "nvidia-tok-location", "", "The NVIDIA .tok file location in the bucket - this must be acquired from NVIDIA and uploaded to your bucket")
	IntVarWithViper(cmd, &o.NvidiaGriddFeatureType, viperNvidiaPrefix, "nvidia-gridd-feature-type", 4, "The gridd feature type - See https://docs.nvidia.com/grid/13.0/grid-licensing-user-guide/index.html#configuring-nls-licensed-client-on-linux for more information")

	cmd.MarkFlagsRequiredTogether("enable-nvidia-support", "nvidia-driver-version", "nvidia-bucket-endpoint", "nvidia-bucket-name", "nvidia-bucket-access", "nvidia-bucket-secret", "nvidia-installer-location", "nvidia-tok-location", "nvidia-gridd-feature-type")
	cmd.MarkFlagsRequiredTogether("cni-version", "crictl-version", "kubernetes-version")
}
