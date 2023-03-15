package trivy

import (
	"time"
)

// Severity is used to parse the value from a report into a programmatic value that can be used for comparisons later.
type Severity string

const (
	NONE     Severity = "NONE"
	LOW      Severity = "LOW"
	MEDIUM   Severity = "MEDIUM"
	HIGH     Severity = "HIGH"
	CRITICAL Severity = "CRITICAL"
)

// CheckSeverity compares two severities to see if a threshold has been met. IE: is sev: HIGH >= check: MEDIUM.
func CheckSeverity(sev, check string) bool {
	var sevValue, checkValue int
	sevValue = parseSeverity(Severity(sev))
	checkValue = parseSeverity(Severity(check))

	if sevValue >= checkValue {
		return true
	}

	return false
}

// parseSeverity takes a Severity and turns it into a numerical value so that it can be compared.
func parseSeverity(val Severity) int {
	switch val {
	case NONE:
		return 1
	case LOW:
		return 2
	case MEDIUM:
		return 3
	case HIGH:
		return 4
	case CRITICAL:
		return 5
	}

	return 0
}

// Report and all its sub-structs is used to unmarshal the json reports into a usable format.
type Report struct {
	SchemaVersion int    `json:"SchemaVersion,omitempty"`
	ArtifactName  string `json:"ArtifactName,omitempty"`
	ArtifactType  string `json:"ArtifactType,omitempty"`
	Metadata      struct {
		Os struct {
			Family string `json:"Family,omitempty"`
			Name   string `json:"Name,omitempty"`
		} `json:"OS,omitempty"`
		ImageConfig struct {
			Architecture string    `json:"architecture,omitempty"`
			Created      time.Time `json:"created,omitempty"`
			Os           string    `json:"os,omitempty"`
			Rootfs       struct {
				Type    string `json:"type,omitempty"`
				DiffIds any    `json:"diff_ids,omitempty"`
			} `json:"rootfs,omitempty"`
			Config struct {
			} `json:"config,omitempty"`
		} `json:"ImageConfig,omitempty"`
	} `json:"Metadata,omitempty"`
	Results []struct {
		Target          string            `json:"Target,omitempty"`
		Class           string            `json:"Class,omitempty"`
		Type            string            `json:"Type,omitempty"`
		Vulnerabilities []Vulnerabilities `json:"Vulnerabilities,omitempty"`
		Secrets         []Secrets         `json:"Secrets,omitempty"`
	} `json:"Results,omitempty"`
}

// CVSS stores all the score data from different sources within the Trivy report.
type CVSS struct {
	Ghsa   *Score `json:"ghsa,omitempty"`
	Nvd    *Score `json:"nvd,omitempty"`
	Redhat *Score `json:"redhat,omitempty"`
}

// Score contains the score values and vectors from a Trivy report.
type Score struct {
	V2Vector string  `json:"V2Vector,omitempty"`
	V3Vector string  `json:"V3Vector,omitempty"`
	V2Score  float64 `json:"V2Score,omitempty"`
	V3Score  float64 `json:"V3Score,omitempty"`
}

// Vulnerabilities contains the vulnerability information from a Trivy report.
type Vulnerabilities struct {
	VulnerabilityID  string `json:"VulnerabilityID,omitempty"`
	PkgID            string `json:"PkgID,omitempty"`
	PkgName          string `json:"PkgName,omitempty"`
	InstalledVersion string `json:"InstalledVersion,omitempty"`
	Layer            struct {
		Digest string `json:",omitempty"`
		DiffID string `json:",omitempty"`
	} `json:"Layer,omitempty"`
	SeveritySource string `json:"Severity,omitempty"`
	PrimaryURL     string `json:"PrimaryURL,omitempty"`
	DataSource     struct {
		ID   string `json:"ID,omitempty"`
		Name string `json:"Name,omitempty"`
		URL  string `json:"URL,omitempty"`
	} `json:"DataSource,omitempty"`
	Title            string    `json:"Title,omitempty"`
	Description      string    `json:"Description,omitempty"`
	Severity         string    `json:"Severity,omitempty"`
	CweIDs           []string  `json:"CweIDs,omitempty"`
	Cvss             CVSS      `json:"CVSS,omitempty"`
	References       []string  `json:"References,omitempty"`
	PublishedDate    time.Time `json:"PublishedDate,omitempty"`
	LastModifiedDate time.Time `json:"LastModifiedDate,omitempty"`
	FixedVersion     string    `json:"FixedVersion,omitempty"`
}

// Secrets contains the secret information from a Trivy report.
type Secrets struct {
	RuleID    string `json:"RuleID,omitempty"`
	Category  string `json:"Category,omitempty"`
	Severity  string `json:"Severity,omitempty"`
	Title     string `json:"Title,omitempty"`
	StartLine int    `json:"StartLine,omitempty"`
	EndLine   int    `json:"EndLine,omitempty"`
	Code      struct {
		Lines []struct {
			Number      int    `json:"Number,omitempty"`
			Content     string `json:"Content,omitempty"`
			IsCause     bool   `json:"IsCause,omitempty"`
			Annotation  string `json:"Annotation,omitempty"`
			Truncated   bool   `json:"Truncated,omitempty"`
			Highlighted string `json:"Highlighted,omitempty"`
			FirstCause  bool   `json:"FirstCause,omitempty"`
			LastCause   bool   `json:"LastCause,omitempty"`
		} `json:"Lines,omitempty"`
	} `json:"Code,omitempty"`
	Match string `json:"Match,omitempty"`
	Layer struct {
		Digest string `json:",omitempty"`
		DiffID string `json:",omitempty"`
	} `json:"Layer,omitempty"`
}
