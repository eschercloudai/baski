package ostack

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type BuildConfig struct {
	ImageName            string `json:"image_name,omitempty"`
	SourceImage          string `json:"source_image,omitempty"`
	Networks             string `json:"networks,omitempty"`
	Flavor               string `json:"flavor,omitempty"`
	AttachConfigDrive    string `json:"attach_config_drive,omitempty"`
	UseFloatingIp        string `json:"use_floating_ip,omitempty"`
	FloatingIpNetwork    string `json:"floating_ip_network,omitempty"`
	CrictlVersion        string `json:"crictl_version,omitempty"`
	KubernetesSemver     string `json:"kubernetes_semver,omitempty"`
	KubernetesRpmVersion string `json:"kubernetes_rpm_version,omitempty"`
	KubernetesSeries     string `json:"kubernetes_series,omitempty"`
	KubernetesDebVersion string `json:"kubernetes_deb_version,omitempty"`
	AnsibleUserVars      string `json:"ansible_user_vars,omitempty"`
}

// ExtractBuildConfig takes a build configuration file and converts it into a BuildConfig
func ExtractBuildConfig(configFilePath string) *BuildConfig {
	log.Println("parsing configuration file")
	var buildconfig *BuildConfig

	configFile, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(configFile, &buildconfig)
	if err != nil {
		log.Fatalln(err)
	}

	return buildconfig
}

// GenerateVariablesFile takes a BuildConfig and converts it into a build configuration file
func GenerateVariablesFile(buildGitDir string, buildConfig *BuildConfig) {
	log.Printf("generating variables file\n")
	outputFileName := strings.Join([]string{"tmp", ".json"}, "")
	outputFile := filepath.Join(buildGitDir, "images/capi/", outputFileName)

	configContent, err := json.Marshal(buildConfig)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(buildConfig)
	fmt.Println(string(configContent))
	err = os.WriteFile(outputFile, configContent, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
