// Package generated provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package generated

// Cvss The name of the package
type Cvss = map[string]interface{}

// CvssType The name of the package
type CvssType = map[string]interface{}

// Health A response to a health request
type Health = string

// ScanResult A scan result.
type ScanResult struct {
	// Cvss The name of the package
	Cvss *Cvss `json:"cvss,omitempty"`

	// Description The name of the package
	Description *string `json:"description,omitempty"`

	// FixedVersion The name of the package
	FixedVersion *string `json:"fixedVersion,omitempty"`

	// InstalledVersion The name of the package
	InstalledVersion *string `json:"installedVersion,omitempty"`

	// PkgName The name of the package
	PkgName *string `json:"pkgName,omitempty"`

	// Severity The name of the package
	Severity *string `json:"severity,omitempty"`

	// VulnerabilityID The name of the package
	VulnerabilityID *string `json:"vulnerabilityID,omitempty"`
}

// ImageID The ID of an image for which to get the scan results for.
type ImageID = string
