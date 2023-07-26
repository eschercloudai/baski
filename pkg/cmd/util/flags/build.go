package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type BuildOptions struct {
	OpenStackFlags
	S3Flags

	Verbose                 bool
	BuildOS                 string
	ImagePrefix             string
	ImageRepo               string
	ImageRepoBranch         string
	CrictlVersion           string
	CniVersion              string
	KubeVersion             string
	ExtraDebs               string
	AdditionalImages        []string
	AddFalco                bool
	AddTrivy                bool
	AddNvidiaSupport        bool
	NvidiaVersion           string
	NvidiaBucket            string
	NvidiaInstallerLocation string
	NvidiaTOKLocation       string
	NvidiaGriddFeatureType  int
}

func (o *BuildOptions) SetOptionsFromViper() {
	o.OpenStackFlags.SetOptionsFromViper()
	o.S3Flags.SetOptionsFromViper()

	// General Flags
	o.Verbose = viper.GetBool(fmt.Sprintf("%s.verbose", viperBuildPrefix))
	o.BuildOS = viper.GetString(fmt.Sprintf("%s.build-os", viperBuildPrefix))
	o.ImagePrefix = viper.GetString(fmt.Sprintf("%s.image-prefix", viperBuildPrefix))
	o.ImageRepo = viper.GetString(fmt.Sprintf("%s.image-repo", viperBuildPrefix))
	o.ImageRepoBranch = viper.GetString(fmt.Sprintf("%s.image-repo-branch", viperBuildPrefix))
	o.CrictlVersion = viper.GetString(fmt.Sprintf("%s.crictl-version", viperBuildPrefix))
	o.CniVersion = viper.GetString(fmt.Sprintf("%s.cni-version", viperBuildPrefix))
	o.KubeVersion = viper.GetString(fmt.Sprintf("%s.kubernetes-version", viperBuildPrefix))
	o.ExtraDebs = viper.GetString(fmt.Sprintf("%s.extra-debs", viperBuildPrefix))
	o.AdditionalImages = viper.GetStringSlice(fmt.Sprintf("%s.additional-images", viperBuildPrefix))
	o.AddFalco = viper.GetBool(fmt.Sprintf("%s.add-falco", viperBuildPrefix))
	o.AddTrivy = viper.GetBool(fmt.Sprintf("%s.add-trivy", viperBuildPrefix))

	// NVIDIA
	o.AddNvidiaSupport = viper.GetBool(fmt.Sprintf("%s.enable-nvidia-support", viperNvidiaPrefix))
	o.NvidiaVersion = viper.GetString(fmt.Sprintf("%s.nvidia-driver-version", viperNvidiaPrefix))
	o.NvidiaBucket = viper.GetString(fmt.Sprintf("%s.nvidia-bucket", viperNvidiaPrefix))
	o.NvidiaInstallerLocation = viper.GetString(fmt.Sprintf("%s.nvidia-installer-location", viperNvidiaPrefix))
	o.NvidiaTOKLocation = viper.GetString(fmt.Sprintf("%s.nvidia-tok-location", viperNvidiaPrefix))
	o.NvidiaGriddFeatureType = viper.GetInt(fmt.Sprintf("%s.nvidia-gridd-feature-type", viperNvidiaPrefix))

}

func (o *BuildOptions) AddFlags(cmd *cobra.Command, imageBuilderRepo string) {
	o.OpenStackFlags.AddFlags(cmd, viperOpenStackPrefix)
	o.S3Flags.AddFlags(cmd, viperS3Prefix)
	// Build flags
	StringVarWithViper(cmd, &o.BuildOS, viperBuildPrefix, "build-os", "ubuntu-2204", "This is the target os to build. Valid values are currently: ubuntu-2004 and ubuntu-2204")
	StringVarWithViper(cmd, &o.ImagePrefix, viperBuildPrefix, "image-prefix", "kube", "This will prefix the image with the value provided. Defaults to 'kube' producing an image name of kube-yymmdd-xxxxxxxx")
	StringVarWithViper(cmd, &o.ImageRepo, viperBuildPrefix, "image-repo", strings.Join([]string{imageBuilderRepo, "git"}, "."), "The imageRepo from which the image builder should be deployed")
	StringVarWithViper(cmd, &o.ImageRepoBranch, viperBuildPrefix, "image-repo-branch", "main", "The branch to checkout from the cloned imageRepo")
	StringVarWithViper(cmd, &o.CniVersion, viperBuildPrefix, "cni-version", "1.2.0", "The CNI plugins version to include to the built image")
	StringVarWithViper(cmd, &o.CrictlVersion, viperBuildPrefix, "crictl-version", "1.25.0", "The crictl-tools version to add to the built image")
	StringVarWithViper(cmd, &o.KubeVersion, viperBuildPrefix, "kubernetes-version", "1.25.3", "The Kubernetes version to add to the built image")
	StringVarWithViper(cmd, &o.ExtraDebs, viperBuildPrefix, "extra-debs", "", "A space-seperated list of any extra (Debian / Ubuntu) packages that should be installed")
	StringSliceVarWithViper(cmd, &o.AdditionalImages, viperBuildPrefix, "additional-images", nil, "Add any additional container images which should be baked into the image")
	BoolVarWithViper(cmd, &o.AddFalco, viperBuildPrefix, "add-falco", false, "If enabled, will install Falco onto the image")
	BoolVarWithViper(cmd, &o.AddTrivy, viperBuildPrefix, "add-trivy", false, "If enabled, will install Trivy onto the image")
	BoolVarWithViper(cmd, &o.Verbose, viperBuildPrefix, "verbose", false, "Enable verbose output to see the information from packer. Not turning this on will mean the process appears to hang while the image build happens")

	// NVIDIA flags
	BoolVarWithViper(cmd, &o.AddNvidiaSupport, viperNvidiaPrefix, "enable-nvidia-support", false, "This will configure NVIDIA support in the image")
	StringVarWithViper(cmd, &o.NvidiaVersion, viperNvidiaPrefix, "nvidia-driver-version", "525.85.05", "The NVIDIA driver version")
	StringVarWithViper(cmd, &o.NvidiaBucket, viperNvidiaPrefix, "nvidia-bucket", "", "The bucket name in which the NVIDIA components are stored")
	StringVarWithViper(cmd, &o.NvidiaInstallerLocation, viperNvidiaPrefix, "nvidia-installer-location", "", "The NVIDIA installer location in the bucket - this must be acquired from NVIDIA and uploaded to your bucket")
	StringVarWithViper(cmd, &o.NvidiaTOKLocation, viperNvidiaPrefix, "nvidia-tok-location", "", "The NVIDIA .tok file location in the bucket - this must be acquired from NVIDIA and uploaded to your bucket")
	IntVarWithViper(cmd, &o.NvidiaGriddFeatureType, viperNvidiaPrefix, "nvidia-gridd-feature-type", 4, "The gridd feature type - See https://docs.nvidia.com/license-system/latest/nvidia-license-system-quick-start-guide/index.html#configuring-nls-licensed-client-on-linux for more information")

	cmd.MarkFlagsRequiredTogether("enable-nvidia-support", "nvidia-driver-version", "nvidia-bucket", "nvidia-installer-location", "nvidia-tok-location", "nvidia-gridd-feature-type")
	cmd.MarkFlagsRequiredTogether("cni-version", "crictl-version", "kubernetes-version")
}
