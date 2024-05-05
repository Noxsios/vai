// Package main is the entry point for the application
package main

import (
	"os"

	"github.com/noxsios/vai/cmd"
)

// main executes the root command.
func main() {
	cli := cmd.NewRootCmd()
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
