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

package constants

import (
	"log"
	"os"
	"reflect"
	"time"
)

type Year struct {
	Months map[string]Month
}

type Month struct {
	Reports map[string]ReportData
}

type ReportData struct {
	Name          string `json:"name"`
	SchemaVersion int    `json:"schema_version"`
	ArtifactName  string `json:"artifact_name"`
	ArtifactType  string `json:"artifact_type"`
	Metadata      struct {
		OS struct {
			Family string `json:"family"`
			Name   string `json:"name"`
		} `json:"os"`

		ImageConfig struct {
			Architecture string    `json:"architecture"`
			Created      time.Time `json:"created"`
			Os           string    `json:"os"`
			Rootfs       struct {
				Type    string      `json:"type"`
				DiffIds interface{} `json:"diff_ids"`
			} `json:"rootfs"`
			Config struct {
			} `json:"config"`
		} `json:"ImageConfig"`
	} `json:"metadata"`

	Results []struct {
		Target          string `json:"Target"`
		Class           string `json:"Class"`
		Type            string `json:"Type,omitempty"`
		Vulnerabilities []struct {
			VulnerabilityID  string `json:"VulnerabilityID"`
			PkgName          string `json:"PkgName"`
			InstalledVersion string `json:"InstalledVersion"`
			Layer            struct {
			} `json:"Layer"`
			SeveritySource string `json:"SeveritySource,omitempty"`
			PrimaryURL     string `json:"PrimaryURL"`
			DataSource     struct {
				ID   string `json:"ID"`
				Name string `json:"Name"`
				URL  string `json:"URL"`
			} `json:"DataSource"`
			Title       string   `json:"Title,omitempty"`
			Description string   `json:"Description"`
			Severity    string   `json:"Severity"`
			CweIDs      []string `json:"CweIDs,omitempty"`
			CVSS        struct {
				Nvd struct {
					V2Vector string  `json:"V2Vector,omitempty"`
					V3Vector string  `json:"V3Vector,omitempty"`
					V2Score  float64 `json:"V2Score,omitempty"`
					V3Score  float64 `json:"V3Score,omitempty"`
				} `json:"nvd,omitempty"`
				Redhat struct {
					V3Vector string  `json:"V3Vector,omitempty"`
					V3Score  float64 `json:"V3Score,omitempty"`
					V2Vector string  `json:"V2Vector,omitempty"`
					V2Score  float64 `json:"V2Score,omitempty"`
				} `json:"redhat,omitempty"`
				Ghsa struct {
					V3Vector string  `json:"V3Vector"`
					V3Score  float64 `json:"V3Score"`
				} `json:"ghsa,omitempty"`
			} `json:"CVSS,omitempty"`
			References       []string  `json:"References"`
			PublishedDate    time.Time `json:"PublishedDate,omitempty"`
			LastModifiedDate time.Time `json:"LastModifiedDate,omitempty"`
			PkgPath          string    `json:"PkgPath,omitempty"`
			FixedVersion     string    `json:"FixedVersion,omitempty"`
		} `json:"Vulnerabilities,omitempty"`
		Secrets []struct {
			RuleID    string `json:"RuleID"`
			Category  string `json:"Category"`
			Severity  string `json:"Severity"`
			Title     string `json:"Title"`
			StartLine int    `json:"StartLine"`
			EndLine   int    `json:"EndLine"`
			Code      struct {
				Lines []struct {
					Number      int    `json:"Number"`
					Content     string `json:"Content"`
					IsCause     bool   `json:"IsCause"`
					Annotation  string `json:"Annotation"`
					Truncated   bool   `json:"Truncated"`
					Highlighted string `json:"Highlighted,omitempty"`
					FirstCause  bool   `json:"FirstCause"`
					LastCause   bool   `json:"LastCause"`
				} `json:"Lines"`
			} `json:"Code"`
			Match   string `json:"Match"`
			Deleted bool   `json:"Deleted"`
			Layer   struct {
			} `json:"Layer"`
		} `json:"Secrets,omitempty"`
	} `json:"Results"`
}

type Env struct {
	AuthURL                  string
	ProjectName              string
	ProjectID                string
	Username                 string
	Password                 string
	Region                   string
	Interface                string
	UserDomainName           string
	ProjectDomainName        string
	IdentityAPIVersion       string
	AuthPlugin               string
	NetworkID                string
	ServerFlavorID           string
	OpenstackBuildConfigPath string
	EnableConfigDrive        string
	ImageRepo                string
	BuildOS                  string
	GhUser                   string
	GhProject                string
	GhToken                  string
	GhPagesBranch            string
}

func CheckEnv(envs *Env, field, envVar string) bool {
	value := reflect.ValueOf(envs).Elem().FieldByName(field)
	if value.String() == "" {
		if key, ok := os.LookupEnv(envVar); ok {
			reflect.ValueOf(envs).Elem().FieldByName(field).SetString(key)
		} else {
			log.Printf("cannot continue without %s\n", envVar)
			return false
		}
	} else {
		err := os.Setenv(envVar, value.String())
		if err != nil {
			log.Printf("couldn't set env var %s.\n", envVar)
			return false
		}
	}
	return true
}

func (e *Env) CheckForEnvVars() {
	canContinue := true
	if !CheckEnv(e, "AuthURL", "OS_AUTH_URL") {
		canContinue = false
	}
	if !CheckEnv(e, "ProjectName", "OS_PROJECT_NAME") {
		canContinue = false
	}
	if !CheckEnv(e, "ProjectID", "OS_PROJECT_ID") {
		canContinue = false
	}
	if !CheckEnv(e, "Username", "OS_USERNAME") {
		canContinue = false
	}
	if !CheckEnv(e, "Password", "OS_PASSWORD") {
		canContinue = false
	}
	if !CheckEnv(e, "Region", "OS_REGION_NAME") {
		canContinue = false
	}
	if !CheckEnv(e, "Interface", "OS_INTERFACE") {
		canContinue = false
	}
	if !CheckEnv(e, "UserDomainName", "OS_USER_DOMAIN_NAME") {
		canContinue = false
	}
	if !CheckEnv(e, "ProjectDomainName", "OS_PROJECT_DOMAIN_NAME") {
		canContinue = false
	}
	if !CheckEnv(e, "IdentityAPIVersion", "OS_IDENTITY_API_VERSION") {
		canContinue = false
	}
	if !CheckEnv(e, "AuthPlugin", "OS_AUTH_PLUGIN") {
		canContinue = false
	}
	if !CheckEnv(e, "NetworkID", "OS_NETWORK_ID") {
		canContinue = false
	}
	if !CheckEnv(e, "ServerFlavorID", "OS_SERVER_FLAVOR_ID") {
		canContinue = false
	}
	if !CheckEnv(e, "EnableConfigDrive", "OS_ENABLE_CONFIG_DRIVE") {
		canContinue = false
	}
	if !CheckEnv(e, "OpenstackBuildConfigPath", "OS_BUILD_CONFIG") {
		canContinue = false
	}
	if !CheckEnv(e, "ImageRepo", "IMAGE_REPO") {
		canContinue = false
	}
	if !CheckEnv(e, "BuildOS", "BUILD_OS") {
		canContinue = false
	}
	if !CheckEnv(e, "GhUser", "GH_USER") {
		canContinue = false
	}
	if !CheckEnv(e, "GhProject", "GH_PROJECT") {
		canContinue = false
	}
	if !CheckEnv(e, "GhToken", "GH_TOKEN") {
		canContinue = false
	}
	if !CheckEnv(e, "GhPagesBranch", "GH_PAGES_BRANCH") {
		canContinue = false
	}

	if !canContinue {
		panic("some required variables are missing - cannot continue")
	}
}

//func (e *Env) SetEnvVarsFromVars() {
//	err := os.Setenv("OS_AUTH_URL", e.AuthURL)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_AUTH_URL")
//	}
//	err = os.Setenv("OS_PROJECT_NAME", e.ProjectName)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_PROJECT_NAME")
//	}
//	err = os.Setenv("OS_PROJECT_ID", e.ProjectID)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_PROJECT_ID")
//	}
//	err = os.Setenv("OS_USERNAME", e.Username)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_USERNAME")
//	}
//	err = os.Setenv("OS_PASSWORD", e.Password)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_PASSWORD")
//	}
//	err = os.Setenv("OS_REGION_NAME", e.Region)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_REGION_NAME")
//	}
//	err = os.Setenv("OS_INTERFACE", e.Interface)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_INTERFACE")
//	}
//	err = os.Setenv("OS_USER_DOMAIN_NAME", e.UserDomainName)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_USER_DOMAIN_NAME")
//	}
//	err = os.Setenv("OS_PROJECT_DOMAIN_ID", e.ProjectDomainName)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_PROJECT_DOMAIN_ID")
//	}
//	err = os.Setenv("OS_IDENTITY_API_VERSION", e.IdentityAPIVersion)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_IDENTITY_API_VERSION")
//	}
//	err = os.Setenv("OS_AUTH_PLUGIN", e.AuthPlugin)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_AUTH_PLUGIN")
//	}
//	err = os.Setenv("OS_NETWORK_ID", e.NetworkID)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_NETWORK_ID")
//	}
//	err = os.Setenv("OS_SERVER_FLAVOR_ID", e.ServerFlavorID)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_SERVER_FLAVOR_ID")
//	}
//	err = os.Setenv("OS_BUILD_CONFIG", e.OpenstackBuildConfigPath)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_BUILD_CONFIG")
//	}
//	err = os.Setenv("OS_ENABLE_CONFIG_DRIVE", e.EnableConfigDrive)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "OS_ENABLE_CONFIG_DRIVE")
//	}
//	err = os.Setenv("IMAGE_REPO", e.ImageRepo)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "IMAGE_REPO")
//	}
//	err = os.Setenv("BUILD_OS", e.BuildOS)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "BUILD_OS")
//	}
//	err = os.Setenv("GH_USER", e.GhUser)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "GH_USER")
//	}
//	err = os.Setenv("GH_PROJECT", e.GhProject)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "GH_PROJECT")
//	}
//	err = os.Setenv("GH_TOKEN", e.GhToken)
//	if err != nil {
//		log.Fatalf("couldn't set env var %s.\n", "GH_TOKEN")
//	}
//}
