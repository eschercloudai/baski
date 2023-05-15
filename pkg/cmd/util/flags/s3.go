package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type S3Flags struct {
	Endpoint  string
	AccessKey string
	SecretKey string
}

func (o *S3Flags) SetOptionsFromViper() {
	o.Endpoint = viper.GetString(fmt.Sprintf("%s.endpoint", viperS3Prefix))
	o.AccessKey = viper.GetString(fmt.Sprintf("%s.access-key", viperS3Prefix))
	o.SecretKey = viper.GetString(fmt.Sprintf("%s.secret-key", viperS3Prefix))

}

func (o *S3Flags) AddFlags(cmd *cobra.Command, imageBuilderRepo string) {
	StringVarWithViper(cmd, &o.Endpoint, viperS3Prefix, "endpoint", "", "The endpoint of the bucket from which to download resources")
	StringVarWithViper(cmd, &o.AccessKey, viperS3Prefix, "access-key", "", "The access key used to access the bucket from which to download resources")
	StringVarWithViper(cmd, &o.SecretKey, viperS3Prefix, "secret-key", "", "The secret key used to access the bucket from which to download resources")

	cmd.MarkFlagsRequiredTogether("endpoint", "access-key", "secret-key")
}
