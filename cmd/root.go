package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	viper.BindEnv("environment")
}

// RootCmd is the main entry point for the Cobra command to run
var RootCmd = &cobra.Command{
	Use:   "epoxy",
	Short: "Epoxy the Proxy of Proxies",
	Long:  "The TCP tunneling solution for the Connection to external proxies",
}
