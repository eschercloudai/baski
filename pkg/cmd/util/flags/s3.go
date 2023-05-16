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
	IsCeph    bool
}

func (o *S3Flags) SetOptionsFromViper() {
	o.Endpoint = viper.GetString(fmt.Sprintf("%s.endpoint", viperS3Prefix))
	o.AccessKey = viper.GetString(fmt.Sprintf("%s.access-key", viperS3Prefix))
	o.SecretKey = viper.GetString(fmt.Sprintf("%s.secret-key", viperS3Prefix))
	o.IsCeph = viper.GetBool(fmt.Sprintf("%s.is-ceph", viperS3Prefix))

}

func (o *S3Flags) AddFlags(cmd *cobra.Command, imageBuilderRepo string) {
	StringVarWithViper(cmd, &o.Endpoint, viperS3Prefix, "endpoint", "", "The endpoint of the bucket from which to download resources")
	StringVarWithViper(cmd, &o.AccessKey, viperS3Prefix, "access-key", "", "The access key used to access the bucket from which to download resources")
	StringVarWithViper(cmd, &o.SecretKey, viperS3Prefix, "secret-key", "", "The secret key used to access the bucket from which to download resources")
	BoolVarWithViper(cmd, &o.IsCeph, viperS3Prefix, "is-ceph", false, "If the S3 endpoint is CEPH then set this to true to allow ansible to work with the endpoint")

	cmd.MarkFlagsRequiredTogether("endpoint", "access-key", "secret-key")
}
