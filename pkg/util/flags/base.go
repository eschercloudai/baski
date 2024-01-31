package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type BaseOptions struct {
	InfraType string
}

func (o *BaseOptions) SetOptionsFromViper() {
	o.InfraType = viper.GetString(fmt.Sprintf("%s.type", viperInfraPrefix))
}

func (o *BaseOptions) AddFlags(cmd *cobra.Command) {
	StringVarWithViper(cmd, &o.InfraType, viperInfraPrefix, "type", "kubevirt", "Targets the settings to use in a config file if supplied or dictates which code runs for a build.")
}
