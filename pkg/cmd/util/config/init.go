package config

import (
	"github.com/spf13/viper"
	"log"
)

// InitConfig will initialise viper and the configuration file.
func InitConfig() {
	viper.SetConfigName("baski")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/tmp/")
	viper.AddConfigPath("/etc/baski/")
	viper.AddConfigPath("$HOME/.baski/")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
	}
}
