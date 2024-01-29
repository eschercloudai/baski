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

package completion

import (
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/spf13/cobra"
	"strings"
)

// CloudCompletionFunc parses clouds.yaml and supplies matching cloud names.
func CloudCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	clouds, err := clientconfig.LoadCloudsYAML()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var matches []string

	for name := range clouds {
		if strings.HasPrefix(name, toComplete) {
			matches = append(matches, name)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
