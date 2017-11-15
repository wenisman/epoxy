package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/wenisman/epoxy/cmd"
)

func init() {
	// set json output for ingestion by log aggregators
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	// default log level should be debug
	log.SetLevel(log.DebugLevel)
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
