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

package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ScanOptions struct {
	BaseOptions
	OpenStackFlags
	KubeVirtFlags
	S3Flags
	ScanSingleOptions
	ScanMultipleOptions

	AutoDeleteImage     bool
	SkipCVECheck        bool
	MaxSeverityScore    float64
	MaxSeverityType     string
	ScanBucket          string
	TrivyignorePath     string
	TrivyignoreFilename string
	TrivyignoreList     []string
}

func (o *ScanOptions) SetOptionsFromViper() {
	o.AutoDeleteImage = viper.GetBool(fmt.Sprintf("%s.auto-delete-image", viperScanPrefix))
	o.SkipCVECheck = viper.GetBool(fmt.Sprintf("%s.skip-cve-check", viperScanPrefix))
	o.MaxSeverityScore = viper.GetFloat64(fmt.Sprintf("%s.max-severity-score", viperScanPrefix))
	o.MaxSeverityType = viper.GetString(fmt.Sprintf("%s.max-severity-type", viperScanPrefix))
	o.ScanBucket = viper.GetString(fmt.Sprintf("%s.scan-bucket", viperScanPrefix))
	o.TrivyignorePath = viper.GetString(fmt.Sprintf("%s.trivyignore-path", viperScanPrefix))
	o.TrivyignoreFilename = viper.GetString(fmt.Sprintf("%s.trivyignore-filename", viperScanPrefix))
	o.TrivyignoreList = viper.GetStringSlice(fmt.Sprintf("%s.trivyignore-list", viperScanPrefix))

	o.BaseOptions.SetOptionsFromViper()
	o.OpenStackFlags.SetOptionsFromViper()
	o.KubeVirtFlags.SetOptionsFromViper()
	o.S3Flags.SetOptionsFromViper()
	o.ScanSingleOptions.SetOptionsFromViper()
	o.ScanMultipleOptions.SetOptionsFromViper()

	// We can override the value of the instance at the scan level
	// This isn't available in the flags as it's already a flag that's available. This is viper only.
	instance := viper.GetString(fmt.Sprintf("%s.flavor-name", viperScanPrefix))
	if instance != "" {
		o.FlavorName = instance
	}
}

func (o *ScanOptions) AddFlags(cmd *cobra.Command) {
	BoolVarWithViper(cmd, &o.AutoDeleteImage, viperScanPrefix, "auto-delete-image", false, "If true, the image will be deleted if a vulnerability check does not succeed - recommended when building new images.")
	BoolVarWithViper(cmd, &o.SkipCVECheck, viperScanPrefix, "skip-cve-check", false, "If true, the image will be allowed even if a vulnerability is detected.")
	Float64VarWithViper(cmd, &o.MaxSeverityScore, viperScanPrefix, "max-severity-score", 7.0, "Can be anything from 0.1 to 10.0. Anything equal to or above this value will cause a failure. (Unless skip-cve-check is supplied)")
	StringVarWithViper(cmd, &o.MaxSeverityType, viperScanPrefix, "max-severity-type", "MEDIUM", "Accepted values are NONE, LOW, MEDIUM, HIGH, CRITICAL. This value will be what the score is checked against For example, a LOW 7.0 would be ignored if the value was HIGH with a `max-severity-score` of 7.0. (Unless skip-cve-check is supplied)")
	StringVarWithViper(cmd, &o.ScanBucket, viperScanPrefix, "scan-bucket", "", "The bucket name to use during scans")
	StringVarWithViper(cmd, &o.TrivyignorePath, viperScanPrefix, "trivyignore-path", "", "The path in the scan-bucket where the trivyignore file is located")
	StringVarWithViper(cmd, &o.TrivyignoreFilename, viperScanPrefix, "trivyignore-filename", "", "The filename of the trivyignore file")
	StringSliceVarWithViper(cmd, &o.TrivyignoreList, viperScanPrefix, "trivyignore-list", []string{}, "A list of CVEs to ignore")

	o.BaseOptions.AddFlags(cmd)
	o.OpenStackFlags.AddFlags(cmd, viperOpenStackPrefix)
	o.KubeVirtFlags.AddFlags(cmd, viperOpenStackPrefix)
	o.S3Flags.AddFlags(cmd)
	o.ScanSingleOptions.AddFlags(cmd)
	o.ScanMultipleOptions.AddFlags(cmd)
}

type ScanSingleOptions struct {
	ImageID string
}

func (o *ScanSingleOptions) SetOptionsFromViper() {
	o.ImageID = viper.GetString(fmt.Sprintf("%s.image-id", viperSinglePrefix))
}

func (o *ScanSingleOptions) AddFlags(cmd *cobra.Command) {
	StringVarWithViper(cmd, &o.ImageID, viperSinglePrefix, "image-id", "", "The ID of the image to scan")
}

type ScanMultipleOptions struct {
	ImageSearch string
	Concurrency int
}

func (o *ScanMultipleOptions) SetOptionsFromViper() {
	o.Concurrency = viper.GetInt(fmt.Sprintf("%s.concurrency", viperMultiplePrefix))
	o.ImageSearch = viper.GetString(fmt.Sprintf("%s.image-search", viperMultiplePrefix))
}

func (o *ScanMultipleOptions) AddFlags(cmd *cobra.Command) {
	IntVarWithViper(cmd, &o.Concurrency, viperMultiplePrefix, "concurrency", 5, "The number of scans that can happen at any one time")
	StringVarWithViper(cmd, &o.ImageSearch, viperMultiplePrefix, "image-search", "", "The prefix of all the images to scan")
}
