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

// KubeVirtFlags are explicitly for QEMU image builds.
type KubeVirtFlags struct {
	QEMUFlags
	StoreInS3      bool
	ImageBucket    string
	ImageName      string
	ImageNamespace string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct.
func (k *KubeVirtFlags) SetOptionsFromViper() {
	k.QEMUFlags.SetOptionsFromViper()
	k.StoreInS3 = viper.GetBool(fmt.Sprintf("%s.store-in-s3", viperKubeVirtPrefix))
	k.ImageBucket = viper.GetString(fmt.Sprintf("%s.image-bucket", viperKubeVirtPrefix))
	k.ImageNamespace = viper.GetString(fmt.Sprintf("%s.image-namespace", viperKubeVirtPrefix))
}

func (k *KubeVirtFlags) AddFlags(cmd *cobra.Command, viperPrefix string) {
	k.QEMUFlags.AddFlags(cmd, viperPrefix)
	BoolVarWithViper(cmd, &k.StoreInS3, viperPrefix, "store-in-s3", false, "Whether to upload the disk image to S3")
	StringVarWithViper(cmd, &k.ImageBucket, viperPrefix, "image-bucket", "10G", "The bucket in S3 to store the image in")
	StringVarWithViper(cmd, &k.ImageNamespace, viperPrefix, "image-namespace", "vm-images", "The Namespace in which to deploy the data volumes for S3 images")
}
