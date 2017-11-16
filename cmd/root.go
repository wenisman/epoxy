package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	// special binding for the environment
	viper.BindEnv("environment")

	// set epoxy environment binding for all commands
	viper.SetEnvPrefix("epoxy")
	viper.SetEnvKeyReplacer(strings.NewReplacer("_", "-"))
	viper.AutomaticEnv()
}

// RootCmd is the main entry point for the Cobra command to run
var RootCmd = &cobra.Command{
	Use:   "epoxy",
	Short: "Epoxy the Proxy of Proxies",
	Long:  "The TCP tunneling solution for the Connection to external proxies",
}
