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

// KubernetesClusterFlags are flags that can be used for the interaction with a kubernetes cluster.
type KubernetesClusterFlags struct {
	KubeconfigPath string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct.
func (k *KubernetesClusterFlags) SetOptionsFromViper() {
	k.KubeconfigPath = viper.GetString(fmt.Sprintf("%s.kubeconfig-path", viperKubernetesClusterPrefix))
}

func (k *KubernetesClusterFlags) AddFlags(cmd *cobra.Command) {
	StringVarWithViper(cmd, &k.KubeconfigPath, viperKubernetesClusterPrefix, "kubeconfig-path", "", "The KubeConfig to use for cluster interaction")
}
