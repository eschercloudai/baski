/*
Copyright 2022 EscherCloud.
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
	"github.com/drew-viles/baskio/cmd/build"
	"github.com/drew-viles/baskio/cmd/publish"
	"github.com/drew-viles/baskio/cmd/scan"
	"github.com/drew-viles/baskio/pkg/constants"
	"github.com/spf13/cobra"
	"log"
)

var (
	rootCmd *cobra.Command
)

// init prepares the tool with all available flag. It also contains the main program loop which runs the tasks.
func init() {
	rootCmd = &cobra.Command{
		Use:   "baskio",
		Short: "Baskio is a tools for building and scanning Kubernetes images within Openstack.",
		Long: `Build And Scan Kubernetes Images on Openstack
This tool has been designed to automatically build images for the Openstack potion of the Kubernetes Image Builder.
It could be extended out to provide images for a variety of other builders however for now it's main goal is to work with Openstack.`,
	}

	buildCmd := build.NewBuildCommand()
	scanCmd := scan.NewScanCommand()
	rootCmd.PersistentFlags().StringVar(&constants.Envs.AuthPlugin, "os-auth-plugin", "password", "The Openstack Auth Plugin")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.AuthURL, "os-auth-url", "", "The Openstack Auth URL")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.IdentityAPIVersion, "os-identity-api-version", "3", "The Openstack Identity API Version")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.Interface, "os-interface", "public", "The Openstack Interface")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.Password, "os-password", "", "The Openstack Password")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.ProjectDomainName, "os-project-domain-name", "default", "The Openstack Project Domain Name")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.ProjectID, "os-project-id", "", "The Openstack Project Name")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.ProjectName, "os-project-name", "", "The Openstack Project Name")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.Region, "os-region-name", "RegionOne", "The Openstack Region Name")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.UserDomainName, "os-user-domain-name", "Default", "The Openstack User Domain Name")
	rootCmd.PersistentFlags().StringVar(&constants.Envs.Username, "os-username", "", "The Openstack UserName")
	requireFlag(rootCmd, "os-auth-url")
	requireFlag(rootCmd, "os-password")
	requireFlag(rootCmd, "os-project-id")
	requireFlag(rootCmd, "os-project-name")
	requireFlag(rootCmd, "os-username")

	commands := []*cobra.Command{
		versionCmd(),
		buildCmd,
		scanCmd,
		publish.NewPublishCommand(),
	}
	rootCmd.AddCommand(commands...)

}

// requireFlag sets flags as required.
func requireFlag(cmd *cobra.Command, flag string) {
	err := cmd.MarkPersistentFlagRequired(flag)
	if err != nil {
		log.Fatalln(err)
	}
}

// Execute runs the execute command for the Cobra library allowing commands & flags to be utilised.
func Execute() error {
	return rootCmd.Execute()
}
