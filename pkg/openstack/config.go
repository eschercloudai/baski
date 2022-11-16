/*
Copyright 2022 EscherCloud.
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
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type OpenstackClouds struct {
	Clouds map[string]OpenstackCloud `yaml:"clouds"`
}

type OpenstackCloud struct {
	Auth               OpenstackAuth `yaml:"auth"`
	RegionName         string        `yaml:"region_name,omitempty"`
	Interface          string        `yaml:"interface,omitempty"`
	IdentityApiVersion int           `yaml:"identity_api_version"`
	AuthType           string        `yaml:"auth_type"`
}

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

type PackerBuildConfig struct {
	ImageName            string `json:"image_name,omitempty"`
	SourceImage          string `json:"source_image,omitempty"`
	Networks             string `json:"networks,omitempty"`
	Flavor               string `json:"flavor,omitempty"`
	AttachConfigDrive    string `json:"attach_config_drive,omitempty"`
	UseFloatingIp        string `json:"use_floating_ip,omitempty"`
	FloatingIpNetwork    string `json:"floating_ip_network,omitempty"`
	CrictlVersion        string `json:"crictl_version,omitempty"`
	ImageVisibility      string `json:"image_visibility,omitempty"`
	KubernetesSemver     string `json:"kubernetes_semver,omitempty"`
	KubernetesRpmVersion string `json:"kubernetes_rpm_version,omitempty"`
	KubernetesSeries     string `json:"kubernetes_series,omitempty"`
	KubernetesDebVersion string `json:"kubernetes_deb_version,omitempty"`
	NodeCustomRolesPost  string `json:"node_custom_roles_post,omitempty"`
	AnsibleUserVars      string `json:"ansible_user_vars,omitempty"`
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
	cloud := c.Clouds[viper.GetString("cloud-name")]
	if cloud.AuthType == "" {
		cloud.AuthType = "password"
	}
	requiredVars := map[string]string{
		"OS_AUTH_URL":             cloud.Auth.AuthURL,
		"OS_PROJECT_NAME":         cloud.Auth.ProjectName,
		"OS_PROJECT_ID":           cloud.Auth.ProjectID,
		"OS_USERNAME":             cloud.Auth.Username,
		"OS_PASSWORD":             cloud.Auth.Password,
		"OS_REGION_NAME":          cloud.RegionName,
		"OS_INTERFACE":            cloud.Interface,
		"OS_USER_DOMAIN_NAME":     cloud.Auth.UserDomainName,
		"OS_PROJECT_DOMAIN_NAME":  "Default",
		"OS_IDENTITY_API_VERSION": strconv.Itoa(cloud.IdentityApiVersion),
		"OS_AUTH_PLUGIN":          cloud.AuthType,
	}

	for k, v := range requiredVars {
		err := os.Setenv(k, v)
		if err != nil {
			log.Fatalln(err)
		}
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
		CrictlVersion:        viper.GetString("build.crictl-version"),
		ImageVisibility:      viper.GetString("build.image-visibility"),
		KubernetesSemver:     "v" + viper.GetString("build.kubernetes-version"),
		KubernetesSeries:     "v" + viper.GetString("build.kubernetes-version"),
		KubernetesRpmVersion: viper.GetString("build.kubernetes-version") + "-0",
		KubernetesDebVersion: viper.GetString("build.kubernetes-version") + "-00",
	}
	if viper.GetBool("build.enable-nvidia-support") {
		buildConfig.NodeCustomRolesPost = "nvidia"
		buildConfig.AnsibleUserVars = fmt.Sprintf("nvidia_installer_url=%s grid_license_server=%s", viper.GetString("build.nvidia-installer-url"), viper.GetString("build.grid-license-server"))
	}
	buildConfig.ImageName = generateImageName(buildConfig.KubernetesSemver)

	return buildConfig
}

// generateImageName creates a name for the image that will be built.
func generateImageName(semVer string) string {
	imageUUID, err := uuid.NewUUID()
	if err != nil {
		log.Fatalln(err)
	}

	buildOS := viper.GetString("build.build-os")

	imageName := buildOS[:3] + buildOS[strings.Index(buildOS, "-")+1:] + "-" + semVer
	if viper.GetBool("build.enable-nvidia-support") {
		imageName = imageName + "-" + "gpu" + "-" + viper.GetString("build.nvidia-driver-version")
	}

	return imageName + "-" + imageUUID.String()[:strings.Index(imageUUID.String(), "-")]
}

// GenerateVariablesFile converts the PackerBuildConfig into a build configuration file that packer can use.
func (p *PackerBuildConfig) GenerateVariablesFile(buildGitDir string) {
	outputFileName := strings.Join([]string{"tmp", ".json"}, "")
	outputFile := filepath.Join(buildGitDir, outputFileName)

	configContent, err := json.Marshal(p)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(configContent))
	err = os.WriteFile(outputFile, configContent, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
