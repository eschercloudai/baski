package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ScanOptions struct {
	OpenStackFlags
	S3Flags

	AddPause            bool
	ImageID             string
	AutoDeleteImage     bool
	SkipCVECheck        bool
	MaxSeverityScore    float64
	MaxSeverityType     string
	TrivyignoreBucket   string
	TrivyignoreFilename string
	TrivyignoreList     []string
}

func (o *ScanOptions) SetOptionsFromViper() {
	o.OpenStackFlags.SetOptionsFromViper()
	o.S3Flags.SetOptionsFromViper()

	// We can override the value of the instance at the scan level
	// This isn't available in the flags as it's already a flag that's available. This is viper only.
	instance := viper.GetString(fmt.Sprintf("%s.flavor-name", viperScanPrefix))
	if instance != "" {
		o.FlavorName = instance
	}
	o.AddPause = viper.GetBool(fmt.Sprintf("%s.add-pause", viperOpenStackPrefix))
	o.ImageID = viper.GetString(fmt.Sprintf("%s.image-id", viperScanPrefix))
	o.AutoDeleteImage = viper.GetBool(fmt.Sprintf("%s.auto-delete-image", viperScanPrefix))
	o.SkipCVECheck = viper.GetBool(fmt.Sprintf("%s.skip-cve-check", viperScanPrefix))
	o.MaxSeverityScore = viper.GetFloat64(fmt.Sprintf("%s.max-severity-score", viperScanPrefix))
	o.MaxSeverityType = viper.GetString(fmt.Sprintf("%s.max-severity-type", viperScanPrefix))
	o.TrivyignoreBucket = viper.GetString(fmt.Sprintf("%s.trivyignore-bucket", viperScanPrefix))
	o.TrivyignoreFilename = viper.GetString(fmt.Sprintf("%s.trivyignore-filename", viperScanPrefix))
	o.TrivyignoreList = viper.GetStringSlice(fmt.Sprintf("%s.trivyignore-list", viperScanPrefix))
}

func (o *ScanOptions) AddFlags(cmd *cobra.Command) {
	o.OpenStackFlags.AddFlags(cmd, viperOpenStackPrefix)
	o.S3Flags.AddFlags(cmd, viperS3Prefix)

	BoolVarWithViper(cmd, &o.AddPause, viperOpenStackPrefix, "add-pause", false, "Adds a pause to the scan step once the server has been created. Only really required for a bug in OpenStack.")
	StringVarWithViper(cmd, &o.ImageID, viperScanPrefix, "image-id", "", "The ID of the image to scan")
	BoolVarWithViper(cmd, &o.AutoDeleteImage, viperScanPrefix, "auto-delete-image", false, "If true, the image will be deleted if a vulnerability check does not succeed - recommended when building new images.")
	BoolVarWithViper(cmd, &o.SkipCVECheck, viperScanPrefix, "skip-cve-check", false, "If true, the image will be allowed even if a vulnerability is detected.")
	Float64VarWithViper(cmd, &o.MaxSeverityScore, viperScanPrefix, "max-severity-score", 7.0, "Can be anything from 0.1 to 10.0. Anything equal to or above this value will cause a failure. (Unless skip-cve-check is supplied)")
	StringVarWithViper(cmd, &o.MaxSeverityType, viperScanPrefix, "max-severity-type", "MEDIUM", "Accepted values are NONE, LOW, MEDIUM, HIGH, CRITICAL. This value will be what the score is checked against For example, a LOW 7.0 would be ignored if the value was HIGH with a `max-severity-score` of 7.0. (Unless skip-cve-check is supplied)")
	StringVarWithViper(cmd, &o.TrivyignoreBucket, viperScanPrefix, "trivyignore-bucket", "", "The bucket name in which the trivyignore file is located")
	StringVarWithViper(cmd, &o.TrivyignoreFilename, viperScanPrefix, "trivyignore-filename", "", "The filename of the trivyignore file")
	StringSliceVarWithViper(cmd, &o.TrivyignoreList, viperScanPrefix, "trivyignore-list", []string{}, "A list of CVEs to ignore")
}
