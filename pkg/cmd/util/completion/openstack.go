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
