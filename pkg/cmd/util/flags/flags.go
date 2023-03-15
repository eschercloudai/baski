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

package flags

import (
	"fmt"
	"github.com/eschercloudai/baski/pkg/cmd/util/completion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GlobalFlags are not specific to a single command and can be included across many
type GlobalFlags struct {
	//BaskiConfigFlag string
	CloudsPathFlag string
	CloudNameFlag  string
}

func (o *GlobalFlags) AddFlags(cmd *cobra.Command) {
	PersistentStringVarWithViper(cmd, &o.CloudsPathFlag, "", "clouds-file", "~/.config/openstack/clouds.yaml", "The location of the openstack clouds.yaml file to use")
	PersistentStringVarWithViper(cmd, &o.CloudNameFlag, "", "cloud-name", "", "The name of the cloud profile to use from the clouds.yaml file")
	//PersistentStringVarWithViper(cmd, &o.BaskiConfigFlag, "", "baski-config", "baski.yaml", "The location of a baski config file")

	if err := cmd.RegisterFlagCompletionFunc("cloud-name", completion.CloudCompletionFunc); err != nil {
		panic(err)
	}

	cmd.MarkFlagsRequiredTogether("clouds-file", "cloud-name")
	//cmd.MarkFlagsMutuallyExclusive("clouds-file", "baski-config")
}

func StringVarWithViper(cmd *cobra.Command, p *string, viperPrefix, name, value, usage string) {
	cmd.Flags().StringVar(p, name, value, usage)
	bindViper(cmd, viperPrefix, name, false)
}

func BoolVarWithViper(cmd *cobra.Command, p *bool, viperPrefix, name string, value bool, usage string) {
	cmd.Flags().BoolVar(p, name, value, usage)
	bindViper(cmd, viperPrefix, name, false)
}

func IntVarWithViper(cmd *cobra.Command, p *int, viperPrefix, name string, value int, usage string) {
	cmd.Flags().IntVar(p, name, value, usage)
	bindViper(cmd, viperPrefix, name, false)
}

func Float64VarWithViper(cmd *cobra.Command, p *float64, viperPrefix, name string, value float64, usage string) {
	cmd.Flags().Float64Var(p, name, value, usage)
	bindViper(cmd, viperPrefix, name, false)
}

func PersistentStringVarWithViper(cmd *cobra.Command, p *string, viperPrefix, name, value, usage string) {
	cmd.PersistentFlags().StringVar(p, name, value, usage)
	bindViper(cmd, viperPrefix, name, true)
}

// bindViper will bind any flag to the equivalent config value
func bindViper(cmd *cobra.Command, prefix, value string, persistent bool) {
	var bindValue string
	if len(prefix) != 0 {
		bindValue = fmt.Sprintf("%s.%s", prefix, value)
	} else {
		bindValue = value
	}

	if persistent {
		err := viper.BindPFlag(bindValue, cmd.PersistentFlags().Lookup(value))
		if err != nil {
			panic(err)
		}

		viper.SetDefault(bindValue, cmd.PersistentFlags().Lookup(value).DefValue)
	} else {

		err := viper.BindPFlag(bindValue, cmd.Flags().Lookup(value))
		if err != nil {
			panic(err)
		}

		viper.SetDefault(bindValue, cmd.Flags().Lookup(value).DefValue)
	}
}
