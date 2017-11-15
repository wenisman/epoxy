package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wenisman/epoxy/lib"
)

var (
	env string
)

func init() {
	pflag.String("environment", "", "The environment that the application is running in")
	pflag.Int("port", 9001, "The port to listen to for inbound connections")

	viper.BindEnv("environment")
	viper.BindPFlags(pflag.CommandLine)

	RootCmd.AddCommand(startCommand)
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the Epoxy server",
	Run: func(cmd *cobra.Command, args []string) {
		lib.LoadConfig()
		lib.SetLogLevel()
		lib.ListenAndServeProxy()
	},
}
