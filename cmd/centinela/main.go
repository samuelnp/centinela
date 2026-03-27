package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"
var stderrWriter io.Writer = os.Stderr
var executeRoot = func() error { return rootCmd.Execute() }
var exitMain = os.Exit

var rootCmd = &cobra.Command{
	Use:     "centinela",
	Short:   "Centinela — development workflow enforcer for Claude/OpenCode projects",
	Version: Version,
}

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Hook integrations (prewrite / postwrite / context / setup / statusline)",
}

func init() {
	rootCmd.AddCommand(hookCmd)
}

func main() {
	if err := executeRoot(); err != nil {
		fmt.Fprintln(stderrWriter, err)
		exitMain(1)
	}
}
