package lib

import (
	"fmt"

	"github.com/spf13/viper"
)

// LoadConfig will read the default and environment specific config files
func LoadConfig() {
	fmt.Println("loading config")
	viper.SetConfigType("yaml")

	// load the default config file
	viper.SetConfigFile("./config/default.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error loading config: %s", err))
	}

	// load the environment specific file
	if viper.InConfig("environment") && viper.GetString("environment") != "" {
		viper.SetConfigFile(fmt.Sprintf("./config/%s.yml", viper.GetString("environment")))
		viper.MergeInConfig()
	}
}
