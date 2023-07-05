/*
Copyright 2023 EscherCloud.

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
	"github.com/eschercloudai/baski/pkg/cmd/build"
	"github.com/eschercloudai/baski/pkg/cmd/publish"
	"github.com/eschercloudai/baski/pkg/cmd/scan"
	"github.com/eschercloudai/baski/pkg/cmd/sign"
	"github.com/eschercloudai/baski/pkg/cmd/util/config"
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
		Long: `Build And Scan Kubernetes Images.
This tool has been designed to automatically build images by leveraging the Kubernetes Image Builder.
As well as building the Kubernetes image it can also be used to scan them once built.`,
	}

	commands := []*cobra.Command{
		versionCmd(),
		build.NewBuildCommand(),
		sign.NewSignCommand(),
		scan.NewScanCommand(),
		publish.NewPublishCommand(),
	}
	cmd.AddCommand(commands...)

}

// Execute runs the execute command for the Cobra library allowing commands & flags to be utilised.
func Execute() error {
	return cmd.Execute()
}
