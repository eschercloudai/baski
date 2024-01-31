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

// QEMUFlags are explicitly for QEMU image builds.
type QEMUFlags struct {
	QemuBinary      string
	DiskSize        string
	OutputDirectory string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct.
func (q *QEMUFlags) SetOptionsFromViper() {
	q.QemuBinary = viper.GetString(fmt.Sprintf("%s.qemu-binary", viperKubeVirtPrefix))
	q.DiskSize = viper.GetString(fmt.Sprintf("%s.disk-size", viperKubeVirtPrefix))
	q.OutputDirectory = viper.GetString(fmt.Sprintf("%s.output-directory", viperKubeVirtPrefix))
}

func (q *QEMUFlags) AddFlags(cmd *cobra.Command, viperPrefix string) {
	StringVarWithViper(cmd, &q.QemuBinary, viperPrefix, "qemu-binary", "qemu-system-x86_64", "The name of the qemu-system-x86_64 binary to use")
	StringVarWithViper(cmd, &q.DiskSize, viperPrefix, "disk-size", "10G", "The size of the VM disk")
	StringVarWithViper(cmd, &q.OutputDirectory, viperPrefix, "output-directory", "/tmp/image-output/", "The directory in which images will be stored once built")
}
