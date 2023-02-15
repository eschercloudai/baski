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

package constants

import (
	"time"
)

var (
	Version     = "v0.0.3-beta.8"
	SupportedOS = []string{
		"ubuntu-2004",
		"ubuntu-2204",
	}
	TrivyVersion = "0.36.1"
)

// Year is used in reports parsing. It is the top level and contains multiple Month(s).
type Year struct {
	Months map[string]Month
}

// Month is used in reports parsing. It is contained within a Year and contains multiple ReportData(s).
type Month struct {
	Reports map[string]ReportData
}

// ReportData is a struct representation of a report that is generated by Trivy.
// It is used for parsing and generating the static sites.
type ReportData struct {
	Name          string `json:"name"`
	ShortName     string `json:"short_name"`
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
