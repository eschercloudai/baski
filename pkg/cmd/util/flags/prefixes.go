package flags

import "fmt"

const (
	viperCloudPrefix   = "cloud"
	viperBuildPrefix   = "build"
	viperScanPrefix    = "scan"
	viperPublishPrefix = "publish"
	viperSignPrefix    = "sign"
)

var (
	viperOpenStackPrefix = fmt.Sprintf("%s.openstack", viperCloudPrefix)
	viperNvidiaPrefix    = fmt.Sprintf("%s.nvidia", viperBuildPrefix)
	viperGithubPrefix    = fmt.Sprintf("%s.github", viperPublishPrefix)
	viperVaultPrefix     = fmt.Sprintf("%s.vault", viperSignPrefix)
	viperGeneratePrefix  = fmt.Sprintf("%s.generate", viperSignPrefix)
)
