package provisoner

import (
	"github.com/drewbernetes/baski/pkg/providers/packer"
	"github.com/drewbernetes/baski/pkg/util/flags"
	"os"
	"time"
)

type BuilderProvisioner interface {
	Init() error
	GeneratePackerConfig() *packer.GlobalBuildConfig
	UpdatePackerBuilders(metadata map[string]string, data []byte) []byte
	PostBuildAction() error
}

// NewBuilder returns a new provisioner based on the infra type that is used for building images.
func NewBuilder(o *flags.BuildOptions) BuilderProvisioner {
	switch o.InfraType {
	case "openstack":
		return newOpenStackBuilder(o)
	case "kubevirt":
		return newKubeVirtBuilder(o)

	}

	return nil
}

type ScannerProvisioner interface {
	Prepare() error
	ScanImages() error
}

func NewScanner(o *flags.ScanOptions) ScannerProvisioner {
	switch o.InfraType {
	case "openstack":
		return newOpenStackScanner(o)
	case "kubevirt":
		return newKubeVirtScanner(o)
	}
	return nil
}

type SignerProvisioner interface {
	SignImage(digest string) error
	ValidateImage(key []byte) error
}

func NewSigner(o *flags.SignOptions) SignerProvisioner {
	switch o.InfraType {
	case "openstack":
		return newOpenStackSigner(o)
	case "kubevirt":
		return newKubeVirtSigner(o)

	}

	return nil
}

// saveImageIDToFile exports the image ID to a file so that it can be read later by the scan system.
func saveImageIDToFile(imgID string) error {
	f, err := os.Create("/tmp/imgid.out")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(imgID))
	if err != nil {
		return err
	}

	return nil
}

// generateBuilderMetadata generates some glance metadata for the image.
func generateBuilderMetadata(o *flags.BuildOptions) map[string]string {
	gpu := "no_gpu"
	if o.AddGpuSupport {
		if o.GpuVendor == "nvidia" {
			gpu = o.NvidiaVersion
		} else if o.GpuVendor == "amd" {
			gpu = o.AMDVersion
		}
	}

	meta := map[string]string{
		"os":         o.BuildOS,
		"k8s":        o.KubeVersion,
		"gpu":        gpu,
		"gpu_vendor": o.GpuVendor,
		"date":       time.Now().Format(time.RFC3339),
	}

	if len(o.AdditionalMetadata) > 0 {
		for k, v := range o.AdditionalMetadata {
			meta[k] = v
		}
	}
	return meta
}
