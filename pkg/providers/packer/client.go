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

package packer

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/drewbernetes/baski/pkg/util/flags"
)

type GlobalBuildConfig struct {
	CniVersion           string            `json:"kubernetes_cni_semver,omitempty"`
	CniDebVersion        string            `json:"kubernetes_cni_deb_version,omitempty"`
	CrictlVersion        string            `json:"crictl_version,omitempty"`
	KubernetesSemver     string            `json:"kubernetes_semver,omitempty"`
	KubernetesRpmVersion string            `json:"kubernetes_rpm_version,omitempty"`
	KubernetesSeries     string            `json:"kubernetes_series,omitempty"`
	KubernetesDebVersion string            `json:"kubernetes_deb_version,omitempty"`
	NodeCustomRolesPre   string            `json:"node_custom_roles_pre,omitempty"`
	NodeCustomRolesPost  string            `json:"node_custom_roles_post,omitempty"`
	AnsibleUserVars      string            `json:"ansible_user_vars,omitempty"`
	ExtraDebs            string            `json:"extra_debs,omitempty"`
	Metadata             map[string]string `json:"-"`
	OpenStackBuildconfig
	KubeVirtBuildConfig
}

func NewCoreBuildconfig(o *flags.BuildOptions) (*GlobalBuildConfig, string) {
	b := &GlobalBuildConfig{
		CniVersion:           "v" + o.CniVersion,
		CniDebVersion:        o.CniDebVersion,
		CrictlVersion:        o.CrictlVersion,
		KubernetesSemver:     "v" + o.KubeVersion,
		KubernetesSeries:     truncateVersion("v" + o.KubeVersion),
		KubernetesRpmVersion: o.KubeRpmVersion,
		KubernetesDebVersion: o.KubeDebVersion,
		ExtraDebs:            o.ExtraDebs,
	}
	var ansibleUserVars string
	var customRoles string
	var additionalImages string
	var securityVars string

	if o.AddGpuSupport {
		customRoles = "gpu"

		if o.GpuVendor == "nvidia" {
			ansibleUserVars = fmt.Sprintf("gpu_vendor=%s nvidia_s3_url=%s nvidia_bucket=%s nvidia_bucket_access=%s nvidia_bucket_secret=%s nvidia_ceph=%t nvidia_installer_location=%s",
				o.GpuVendor,
				o.Endpoint,
				o.NvidiaBucket,
				o.AccessKey,
				o.SecretKey,
				o.IsCeph,
				o.NvidiaInstallerLocation)

			if o.NvidiaTOKLocation != "" {
				ansibleUserVars = fmt.Sprintf("%s nvidia_tok_location=%s",
					ansibleUserVars,
					o.NvidiaTOKLocation)
			}

			if o.NvidiaGriddFeatureType != -1 {
				ansibleUserVars = fmt.Sprintf("%s gridd_feature_type=%d",
					ansibleUserVars,
					o.NvidiaGriddFeatureType)
			}
		} else if o.GpuVendor == "amd" {
			ansibleUserVars = fmt.Sprintf("gpu_vendor=%s amd_version=%s amd_deb_version=%s gpu_amd_usecase=%s",
				o.GpuVendor,
				o.AMDVersion,
				o.AMDDebVersion,
				o.AMDUseCase)
		}
	}

	// Little workaround for people leaving an empty field or not having the field in the yaml.
	// viper likes to replace a non-existent entry with the string "[]" even when the default is nil.
	if o.AdditionalImages != nil {
		if len(o.AdditionalImages) > 0 {
			if o.AdditionalImages[0] == "[]" {
				o.AdditionalImages = nil
			}
		}
	}
	if o.AdditionalImages != nil {
		for k, v := range o.AdditionalImages {
			if k == 0 {
				additionalImages = additionalImages + v
			} else {
				additionalImages = additionalImages + "," + v
			}
		}
		if len(ansibleUserVars) == 0 {
			ansibleUserVars = "load_additional_components=true additional_registry_images=true additional_registry_images_list=" + additionalImages
		} else {
			ansibleUserVars = ansibleUserVars + " load_additional_components=true additional_registry_images=true additional_registry_images_list=" + additionalImages
		}
	}

	if o.AddFalco || o.AddTrivy {
		if len(customRoles) == 0 {
			customRoles = "security"
		} else {
			customRoles = customRoles + " security"
		}

		if o.AddFalco && !o.AddTrivy {
			securityVars = "install_falco=true"
		} else if !o.AddFalco && o.AddTrivy {
			securityVars = "install_trivy=true"
		} else {
			securityVars = "install_falco=true install_trivy=true"
		}
		if len(ansibleUserVars) == 0 {
			ansibleUserVars = securityVars
		} else {
			ansibleUserVars = ansibleUserVars + " " + securityVars
		}
	}

	b.NodeCustomRolesPre = customRoles
	b.AnsibleUserVars = ansibleUserVars

	return b, generateImageName(o.ImagePrefix)
}

type BuildersModifier struct {
	Function func(metadata map[string]string, data []byte) []byte
	Metadata map[string]string
}

// UpdatePackerBuildersJson pre-populates the metadata field in the packer.json file as objects cannot be passed as variables in packer.
func UpdatePackerBuildersJson(dir string, infra string, modifier BuildersModifier) error {
	// change infra to qemu if kubevirt is the infra type as this is what is needed to build
	if infra == "kubevirt" {
		infra = "qemu"
	}

	file, err := os.OpenFile(filepath.Join(dir, "images", "capi", "packer", infra, "packer.json"), os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	res := modifier.Function(modifier.Metadata, data)

	if res == nil {
		return nil
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(res)
	if err != nil {
		return err
	}
	return nil
}

// GenerateVariablesFile converts the GlobalBuildConfig into a build configuration file that packer can use.
func (p *GlobalBuildConfig) GenerateVariablesFile(buildGitDir string) {
	outputFileName := strings.Join([]string{"tmp", ".json"}, "")
	outputFile := filepath.Join(buildGitDir, outputFileName)

	configContent, err := json.Marshal(p)
	if err != nil {
		log.Fatalln(err)
	}

	err = os.WriteFile(outputFile, configContent, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

// Extract Kubernetes series with the assumption we want vX.XX
func truncateVersion(version string) string {
	re := regexp.MustCompile(`v\d+\.\d+`)
	return re.FindString(version)
}

// generateImageName creates a name for the image that will be built.
func generateImageName(imagePrefix string) string {
	imageUUID, err := uuid.NewRandom()
	if err != nil {
		log.Fatalln(err)
	}

	shortDate := time.Now().Format("060102")
	shortUUID := imageUUID.String()[:strings.Index(imageUUID.String(), "-")]

	return imagePrefix + "-" + shortDate + "-" + shortUUID
}
