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
		Short: "Baski is a tools for building and scanning Kubernetes images within Openstack.",
		Long: `Build And Scan Kubernetes Images on Openstack
This tool has been designed to automatically build images for the Openstack potion of the Kubernetes Image Builder.
It could be extended out to provide images for a variety of other builders however for now it's main goal is to work with Openstack.`,
	}

	//cmd.PersistentFlags().StringVar(&flags.cloudsPathFlag, "clouds-file", "~/.config/openstack/clouds.yaml", "The location of the openstack clouds.yaml file to use")
	//cmd.PersistentFlags().StringVar(&flags.cloudNameFlag, "cloud-name", "", "The name of the cloud profile to use from the clouds.yaml file")
	//
	//if err := cmd.RegisterFlagCompletionFunc("cloud-name", completion.CloudCompletionFunc); err != nil {
	//	panic(err)
	//}
	//cmd.PersistentFlags().StringVar(&flags.baskiConfigFlag, "baski-config", "baski.yaml", "The location of a baski config file")
	//
	//bindPersistentViper(cmd, "clouds-file")
	//bindPersistentViper(cmd, "cloud-name")
	//bindPersistentViper(cmd, "baski-config")
	//
	//cmd.MarkFlagsRequiredTogether("clouds-file", "cloud-name")
	//cmd.MarkFlagsMutuallyExclusive("clouds-file", "baski-config")

	commands := []*cobra.Command{
		versionCmd(),
		build.NewBuildCommand(),
		sign.NewSignCommand(),
		scan.NewScanCommand(),
		publish.NewPublishCommand(),
	}
	cmd.AddCommand(commands...)

}

//
//// initConfig will initialise viper and the configuration file.
//func initConfig() {
//	if flags.baskiConfigFlag != "" {
//		viper.SetConfigFile(flags.baskiConfigFlag)
//	} else {
//		viper.SetConfigName("baski")
//		viper.SetConfigType("yaml")
//		viper.AddConfigPath(".")
//
//		err := viper.ReadInConfig()
//		if err != nil {
//			panic(fmt.Errorf("fatal error config file: %w", err))
//		}
//	}
//
//	if err := viper.ReadInConfig(); err != nil {
//		log.Println(err)
//	}
//}

//// bindViper will bind any flag and envvar to the config
//func bindViper(cmd *cobra.Command, bindValue, flagValue string) {
//	err := viper.BindPFlag(bindValue, cmd.Flags().Lookup(flagValue))
//	if err != nil {
//		panic(err)
//	}
//
//	viper.SetDefault(bindValue, cmd.Flags().Lookup(flagValue).DefValue)
//}
//
//// bindPersistentViper will bind any persistent flag and envvar to the config
//func bindPersistentViper(cmd *cobra.Command, value string) {
//	err := viper.BindPFlag(value, cmd.PersistentFlags().Lookup(value))
//	if err != nil {
//		panic(err)
//	}
//
//	viper.SetDefault(value, cmd.PersistentFlags().Lookup(value).DefValue)
//}

// Execute runs the execute command for the Cobra library allowing commands & flags to be utilised.
func Execute() error {
	return cmd.Execute()
}
