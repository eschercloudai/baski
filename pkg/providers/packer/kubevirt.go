package packer

// KubeVirtBuildConfig adds additional packer vars for Kubevirt
type KubeVirtBuildConfig struct {
	QemuBinary      string `json:"qemu_binary"`
	DiskSize        string `json:"disk_size"`
	OutputDirectory string `json:"output_directory"`
}
