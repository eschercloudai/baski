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

package cmd

import (
	"github.com/drewbernetes/baski/pkg/cmd/build"
	"github.com/drewbernetes/baski/pkg/cmd/scan"
	"github.com/drewbernetes/baski/pkg/cmd/sign"
	"github.com/drewbernetes/baski/pkg/cmd/util/config"
	"github.com/spf13/cobra"
)

var (
	cmd *cobra.Command
)

// init prepares the tool with all available flag. It also contains the main program loop which runs the tasks.
func init() {
	cobra.OnInitialize(config.InitConfig)

	cmd = &cobra.Command{
		Use:   "baski",
		Short: "Baski is a tools for building and scanning Kubernetes images.",
		Long: `Build And Scan Kubernetes Images 
This tool has been designed to automatically build images for the Openstack potion of the Kubernetes Image Builder.
It could be extended out to provide images for a variety of other builders however for now it's main goal is to work with Openstack.`,
	}

	commands := []*cobra.Command{
		versionCmd(),
		build.NewBuildCommand(),
		sign.NewSignCommand(),
		scan.NewScanCommand(),
	}
	cmd.AddCommand(commands...)

}

// Execute runs the execute command for the Cobra library allowing commands & flags to be utilised.
func Execute() error {
	return cmd.Execute()
}
