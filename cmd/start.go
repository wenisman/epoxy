package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wenisman/epoxy/lib"
)

var (
	env string
)

func init() {
	startCommand.Flags().StringVar(&env, "environment", "", "the environment that Epoxy is deployed to")
	viper.BindPFlag("environment", RootCmd.Flags().Lookup("environment"))

	RootCmd.AddCommand(startCommand)
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the Epoxy server",
	Run: func(cmd *cobra.Command, args []string) {
		lib.LoadConfig()
		lib.ListenAndServeProxy()
	},
}
