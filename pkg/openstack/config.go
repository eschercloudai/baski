package ostack

import (
	"encoding/json"
	"log"
	"os"
)

type BuildConfig struct {
	BuildName            string `json:"build_name"`
	DistroName           string `json:"distro_name"`
	GuestOsType          string `json:"guest_os_type"`
	OsDisplayName        string `json:"os_display_name"`
	ShutdownCommand      string `json:"shutdown_command"`
	SshUsername          string `json:"ssh_username"`
	SourceImage          string `json:"source_image"`
	Networks             string `json:"networks"`
	Flavor               string `json:"flavor"`
	AttachConfigDrive    string `json:"attach_config_drive"`
	UseFloatingIp        string `json:"use_floating_ip"`
	FloatingIpNetwork    string `json:"floating_ip_network"`
	CrictlVersion        string `json:"crictl_version"`
	KubernetesSemver     string `json:"kubernetes_semver"`
	KubernetesRpmVersion string `json:"kubernetes_rpm_version"`
	KubernetesSeries     string `json:"kubernetes_series"`
	KubernetesDebVersion string `json:"kubernetes_deb_version"`
}

// ParseBuildConfig takes a build configuration file and converts it into a BuildConfig
func ParseBuildConfig(configFilePath string) *BuildConfig {
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
