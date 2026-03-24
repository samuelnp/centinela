package main

import "github.com/spf13/cobra"

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run project migrations",
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
