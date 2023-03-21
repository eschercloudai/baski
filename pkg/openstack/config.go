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

package ostack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// OpenstackClouds exists to contain the contents of the clouds.yaml file for Openstack
type OpenstackClouds struct {
	Clouds map[string]OpenstackCloud `yaml:"clouds"`
}

// OpenstackCloud is a singular cloud definition in the clouds.yaml file for Openstack.
type OpenstackCloud struct {
	Auth               OpenstackAuth `yaml:"auth"`
	RegionName         string        `yaml:"region_name,omitempty"`
	Interface          string        `yaml:"interface,omitempty"`
	IdentityApiVersion int           `yaml:"identity_api_version"`
	AuthType           string        `yaml:"auth_type"`
}

// OpenstackAuth is the auth section of a singular cloud in the clouds.yaml file for Openstack.
type OpenstackAuth struct {
	AuthURL                     string `yaml:"auth_url"`
	Username                    string `yaml:"username,omitempty"`
	Password                    string `yaml:"password,omitempty"`
	ApplicationCredentialID     string `yaml:"application_credential_id,omitempty"`
	ApplicationCredentialSecret string `yaml:"application_credential_secret,omitempty"`
	ProjectID                   string `yaml:"project_id"`
	ProjectName                 string `yaml:"project_name"`
	UserDomainName              string `yaml:"user_domain_name"`
}

// PackerBuildConfig exists to allow variables to be parsed into a packer json file which can then be used for a build.
type PackerBuildConfig struct {
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
	NodeCustomRolesPost  string `json:"node_custom_roles_post,omitempty"`
	AnsibleUserVars      string `json:"ansible_user_vars,omitempty"`
	ExtraDebs            string `json:"extra_debs,omitempty"`
	VolumeType           string `json:"volume_type"`
	ImageDiskFormat      string `json:"image_disk_format"`
}

// InitOpenstack translates the clouds.yaml file into a struct to be used in app.
func InitOpenstack() (cloudsConfig *OpenstackClouds) {
	return parseCloudsConfig(viper.GetString("clouds-file"))
}

// InitPackerConfig translates all the inputs into the global struct so that it can be utilised as required.
func InitPackerConfig() (packerBuildConfig *PackerBuildConfig) {
	return buildConfigFromInputs()
}

// parseCloudsConfig will read the contents of the clouds.yaml file for Openstack and parse it into a OpenstackClouds struct.
func parseCloudsConfig(cloudsPath string) *OpenstackClouds {
	var cloudsConfig *OpenstackClouds

	if strings.Split(cloudsPath, "/")[0] == "~" {
		prefix, err := os.UserHomeDir()
		if err != nil {
			log.Fatalln(err)
		}
		cloudsPath = filepath.Join(prefix, filepath.Join(strings.Split(cloudsPath, "/")[1:]...))
	}

	config, err := os.ReadFile(cloudsPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(config, &cloudsConfig)
	if err != nil {
		panic(err)
	}

	return cloudsConfig
}

// SetOpenstackEnvs sets the environment variables for the build command to be able to connect to Openstack.
func (c *OpenstackClouds) SetOpenstackEnvs() {
	err := os.Setenv("OS_CLOUD", viper.GetString("cloud-name"))
	if err != nil {
		log.Fatalln(err)
	}
}

// buildConfigFromInputs takes the application inputs and converts it into a PackerBuildConfig.
func buildConfigFromInputs() *PackerBuildConfig {
	buildConfig := &PackerBuildConfig{
		SourceImage:          viper.GetString("build.source-image"),
		Networks:             viper.GetString("build.network-id"),
		Flavor:               viper.GetString("build.flavor-name"),
		AttachConfigDrive:    strconv.FormatBool(viper.GetBool("build.attach-config-drive")),
		UseFloatingIp:        strconv.FormatBool(viper.GetBool("build.use-floating-ip")),
		FloatingIpNetwork:    viper.GetString("build.floating-ip-network-name"),
		CniVersion:           "v" + viper.GetString("build.cni-version"),
		CniDebVersion:        viper.GetString("build.cni-version") + "-00",
		CrictlVersion:        viper.GetString("build.crictl-version"),
		ImageVisibility:      viper.GetString("build.image-visibility"),
		KubernetesSemver:     "v" + viper.GetString("build.kubernetes-version"),
		KubernetesSeries:     "v" + viper.GetString("build.kubernetes-version"),
		KubernetesRpmVersion: viper.GetString("build.kubernetes-version") + "-0",
		KubernetesDebVersion: viper.GetString("build.kubernetes-version") + "-00",
		ExtraDebs:            viper.GetString("build.extra-debs"),
		ImageDiskFormat:      viper.GetString("build.image-disk-format"),
		VolumeType:           viper.GetString("build.volume-type"),
	}
	if viper.GetBool("build.enable-nvidia-support") {
		buildConfig.NodeCustomRolesPost = "nvidia"
		buildConfig.AnsibleUserVars = fmt.Sprintf("nvidia_s3_url=%s nvidia_bucket=%s nvidia_bucket_access=%s nvidia_bucket_secret=%s nvidia_installer_location=%s nvidia_tok_location=%s gridd_feature_type=%s",
			viper.GetString("build.nvidia-bucket-endpoint"),
			viper.GetString("build.nvidia-bucket-name"),
			viper.GetString("build.nvidia-bucket-access"),
			viper.GetString("build.nvidia-bucket-secret"),
			viper.GetString("build.nvidia-installer-location"),
			viper.GetString("build.nvidia-tok-location"),
			viper.GetString("build.gridd-feature-type"))
	}
	buildConfig.ImageName = generateImageName()

	return buildConfig
}

// generateImageName creates a name for the image that will be built.
func generateImageName() string {
	imageUUID, err := uuid.NewRandom()
	if err != nil {
		log.Fatalln(err)
	}

	shortDate := time.Now().Format("060102")
	shortUUID := imageUUID.String()[:strings.Index(imageUUID.String(), "-")]

	return viper.GetString("build.image-prefix") + "-" + shortDate + "-" + shortUUID
}

// GenerateBuilderMetadata generates some glance metadata for the image.
func GenerateBuilderMetadata() map[string]string {
	gpu := "no_gpu"
	if viper.GetBool("build.enable-nvidia-support") {
		gpu = viper.GetString("build.nvidia-driver-version")
	}
	return map[string]string{
		"os":          viper.GetString("build.build-os"),
		"k8s":         viper.GetString("build.kubernetes-version"),
		"gpu":         gpu,
		"date":        time.RFC3339,
		"rootfs_uuid": viper.GetString("build.rootfs-uuid"),
	}
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

// GenerateVariablesFile converts the PackerBuildConfig into a build configuration file that packer can use.
func (p *PackerBuildConfig) GenerateVariablesFile(buildGitDir string) {
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
