package main

import (
	"fmt"
	"os"

	"github.com/wenisman/epoxy/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
		// do some work
		//lib.ListenAndServeProxy()
	}
}
