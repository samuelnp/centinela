package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "centinela",
	Short:   "Centinela — development workflow enforcer for Claude Code projects",
	Version: Version,
}

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Claude Code hook integrations (prewrite / postwrite / context)",
}

func init() {
	rootCmd.AddCommand(hookCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
