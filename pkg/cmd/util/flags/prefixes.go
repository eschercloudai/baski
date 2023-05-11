package flags

import "fmt"

const (
	viperCloudPrefix   = "cloud"
	viperS3Prefix      = "s3"
	viperBuildPrefix   = "build"
	viperScanPrefix    = "scan"
	viperSignPrefix    = "sign"
	viperPublishPrefix = "publish"
)

var (
	viperOpenStackPrefix = fmt.Sprintf("%s.openstack", viperCloudPrefix)
	viperNvidiaPrefix    = fmt.Sprintf("%s.nvidia", viperBuildPrefix)
	viperGithubPrefix    = fmt.Sprintf("%s.github", viperPublishPrefix)
	viperVaultPrefix     = fmt.Sprintf("%s.vault", viperSignPrefix)
	viperGeneratePrefix  = fmt.Sprintf("%s.generate", viperSignPrefix)
)
