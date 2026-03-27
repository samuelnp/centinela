package main

import "github.com/spf13/cobra"

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate and validate human-readable project documentation",
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
