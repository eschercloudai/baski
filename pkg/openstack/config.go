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
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
	NodeCustomRolesPre   string `json:"node_custom_roles_pre,omitempty"`
	NodeCustomRolesPost  string `json:"node_custom_roles_post,omitempty"`
	AnsibleUserVars      string `json:"ansible_user_vars,omitempty"`
	ExtraDebs            string `json:"extra_debs,omitempty"`
	VolumeType           string `json:"volume_type"`
	ImageDiskFormat      string `json:"image_disk_format"`
}

// InitOpenstack will read the contents of the clouds.yaml file for Openstack and parse it into a OpenstackClouds struct.
func InitOpenstack(cloudsFile string) *OpenstackClouds {
	var cloudsConfig *OpenstackClouds

	if strings.Split(cloudsFile, "/")[0] == "~" {
		prefix, err := os.UserHomeDir()
		if err != nil {
			log.Fatalln(err)
		}
		cloudsFile = filepath.Join(prefix, filepath.Join(strings.Split(cloudsFile, "/")[1:]...))
	}

	config, err := os.ReadFile(cloudsFile)
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
func (c *OpenstackClouds) SetOpenstackEnvs(cloudName string) {
	err := os.Setenv("OS_CLOUD", cloudName)
	if err != nil {
		log.Fatalln(err)
	}
}

// InitPackerConfig takes the application inputs and converts it into a PackerBuildConfig.
func InitPackerConfig(o *flags.BuildOptions) *PackerBuildConfig {
	buildConfig := &PackerBuildConfig{
		SourceImage:          o.SourceImageID,
		Networks:             o.NetworkID,
		Flavor:               o.FlavorName,
		AttachConfigDrive:    strconv.FormatBool(o.AttachConfigDrive),
		UseFloatingIp:        strconv.FormatBool(o.UseFloatingIP),
		FloatingIpNetwork:    o.FloatingIPNetworkName,
		CniVersion:           "v" + o.CniVersion,
		CniDebVersion:        o.CniVersion + "-00",
		CrictlVersion:        o.CrictlVersion,
		ImageVisibility:      o.ImageVisibility,
		KubernetesSemver:     "v" + o.KubeVersion,
		KubernetesSeries:     "v" + o.KubeVersion,
		KubernetesRpmVersion: o.KubeVersion + "-0",
		KubernetesDebVersion: o.KubeVersion + "-00",
		ExtraDebs:            o.ExtraDebs,
		ImageDiskFormat:      o.ImageDiskFormat,
		VolumeType:           o.VolumeType,
	}

	var ansibleUserVars string
	var securityVars string
	var customRoles string

	if o.AddNvidiaSupport {
		customRoles = "nvidia"
		ansibleUserVars = fmt.Sprintf("nvidia_s3_url=%s nvidia_bucket=%s nvidia_bucket_access=%s nvidia_bucket_secret=%s nvidia_installer_location=%s nvidia_tok_location=%s gridd_feature_type=%d",
			o.Endpoint,
			o.NvidiaBucket,
			o.AccessKey,
			o.SecretKey,
			o.NvidiaInstallerLocation,
			o.NvidiaTOKLocation,
			o.NvidiaGriddFeatureType)
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

// GenerateBuilderMetadata generates some glance metadata for the image.
func GenerateBuilderMetadata(o *flags.BuildOptions) map[string]string {
	gpu := "no_gpu"
	if o.AddNvidiaSupport {
		gpu = o.NvidiaVersion
	}
	return map[string]string{
		"os":          o.BuildOS,
		"k8s":         o.KubeVersion,
		"gpu":         gpu,
		"date":        time.RFC3339,
		"rootfs_uuid": o.RootfsUUID,
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
