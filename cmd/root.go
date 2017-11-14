package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd is the main entry point for the Cobra command to run
var RootCmd = &cobra.Command{
	Use:   "epoxy",
	Short: "Epoxy the Proxy of Proxies",
	Long:  "The TCP tunneling solution for the Connection to external proxies",
}
