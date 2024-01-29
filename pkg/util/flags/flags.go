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

func StringVarWithViper(cmd *cobra.Command, p *string, viperPrefix, name, value, usage string) {
	cmd.Flags().StringVar(p, name, value, usage)
	bindViper(cmd, viperPrefix, name, false)
}

func StringSliceVarWithViper(cmd *cobra.Command, p *[]string, viperPrefix, name string, value []string, usage string) {
	cmd.Flags().StringSliceVar(p, name, value, usage)
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

func Int64VarWithViper(cmd *cobra.Command, p *int64, viperPrefix, name string, value int64, usage string) {
	cmd.Flags().Int64Var(p, name, value, usage)
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
