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
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/drewbernetes/baski/pkg/util/flags"

	"github.com/google/uuid"
)

// Buildconfig exists to allow variables to be parsed into a packer json file which can then be used for a build.
type Buildconfig struct {
	ImageName            string `json:"image_name,omitempty"`
	SourceImage          string `json:"source_image"`
	Networks             string `json:"networks"`
	Flavor               string `json:"flavor"`
	AttachConfigDrive    string `json:"attach_config_drive,omitempty"`
	UseFloatingIp        string `json:"use_floating_ip,omitempty"`
	FloatingIpNetwork    string `json:"floating_ip_network,omitempty"`
	CniVersion           string `json:"kubernetes_cni_semver,omitempty"`
	CniDebVersion        string `json:"kubernetes_cni_deb_version,omitempty"`
	CrictlVersion        string `json:"crictl_version,omitempty"`
	ImageVisibility      string `json:"image_visibility,omitempty"`
	KubernetesSemver     string `json:"kubernetes_semver,omitempty"`
	KubernetesRpmVersion string `json:"kubernetes_rpm_version,omitempty"`
	KubernetesSeries     string `json:"kubernetes_series,omitempty"`
	KubernetesDebVersion string `json:"kubernetes_deb_version,omitempty"`
	NodeCustomRolesPre   string `json:"node_custom_roles_pre,omitempty"`
	NodeCustomRolesPost  string `json:"node_custom_roles_post,omitempty"`
	AnsibleUserVars      string `json:"ansible_user_vars,omitempty"`
	ExtraDebs            string `json:"extra_debs,omitempty"`
	ImageDiskFormat      string `json:"image_disk_format"`
	VolumeType           string `json:"volume_type"`
	VolumeSize           string `json:"volume_size"`
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

// addMetadataToBuilders inserts the metadata into the packer's builder section.
func addMetadataToBuilders(metadata map[string]string, data []byte) []byte {
	jsonStruct := struct {
		Builders       []map[string]interface{} `json:"builders"`
		PostProcessors []map[string]interface{} `json:"post-processors"`
		Provisioners   []map[string]interface{} `json:"provisioners"`
		Variables      map[string]interface{}   `json:"variables"`
	}{}

	err := json.Unmarshal(data, &jsonStruct)
	if err != nil {
		log.Fatalln(err)
	}

	jsonStruct.Builders[0]["metadata"] = metadata

	res, err := json.Marshal(jsonStruct)
	if err != nil {
		log.Fatalln(err)
	}

	return res
}

// UpdatePackerBuildersJson pre-populates the metadata field in the packer.json file as objects cannot be passed as variables in packer.
func UpdatePackerBuildersJson(dir string, metadata map[string]string) error {
	file, err := os.OpenFile(filepath.Join(dir, "images", "capi", "packer", "openstack", "packer.json"), os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	res := addMetadataToBuilders(metadata, data)

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

// InitConfig takes the application inputs and converts it into a Buildconfig.
func InitConfig(o *flags.BuildOptions) *Buildconfig {
	buildConfig := &Buildconfig{
		SourceImage:          o.SourceImageID,
		Networks:             o.NetworkID,
		Flavor:               o.FlavorName,
		AttachConfigDrive:    strconv.FormatBool(o.AttachConfigDrive),
		UseFloatingIp:        strconv.FormatBool(o.UseFloatingIP),
		FloatingIpNetwork:    o.FloatingIPNetworkName,
		CniVersion:           "v" + o.CniVersion,
		CniDebVersion:        o.CniDebVersion,
		CrictlVersion:        o.CrictlVersion,
		ImageVisibility:      o.ImageVisibility,
		KubernetesSemver:     "v" + o.KubeVersion,
		KubernetesSeries:     truncateVersion("v" + o.KubeVersion),
		KubernetesRpmVersion: o.KubeRpmVersion,
		KubernetesDebVersion: o.KubeDebVersion,
		ExtraDebs:            o.ExtraDebs,
		ImageDiskFormat:      o.ImageDiskFormat,
		VolumeType:           o.VolumeType,
		VolumeSize:           strconv.Itoa(o.VolumeSize),
	}

	var ansibleUserVars string
	var additionalImages string
	var securityVars string
	var customRoles string

	// Little workaround for people leaving an empty field or not having the field in the yaml.
	// viper likes to replace a non-existent entry with the string "[]" even when the default is nil.
	if o.AdditionalImages != nil {
		if len(o.AdditionalImages) > 0 {
			if o.AdditionalImages[0] == "[]" {
				o.AdditionalImages = nil
			}
		}
	}

	if o.AddNvidiaSupport {
		customRoles = "nvidia"

		ansibleUserVars = fmt.Sprintf("nvidia_s3_url=%s nvidia_bucket=%s nvidia_bucket_access=%s nvidia_bucket_secret=%s nvidia_ceph=%t nvidia_installer_location=%s",
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

	buildConfig.NodeCustomRolesPre = customRoles
	buildConfig.AnsibleUserVars = ansibleUserVars
	buildConfig.ImageName = generateImageName(o.ImagePrefix)

	return buildConfig
}

// GenerateVariablesFile converts the Buildconfig into a build configuration file that packer can use.
func (p *Buildconfig) GenerateVariablesFile(buildGitDir string) {
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
