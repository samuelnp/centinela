package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/migration"
	"github.com/samuelnp/centinela/internal/ui"
)

var hookMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Hook: detect managed docs requiring migration",
	RunE:  runHookMigrate,
}

func init() {
	hookCmd.AddCommand(hookMigrateCmd)
}

func runHookMigrate(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck // drain stdin to avoid SIGPIPE
	if _, err := os.Stat("PROJECT.md.template"); err != nil {
		return nil
	}
	plan, err := migration.BuildPlan(".")
	if err != nil || !plan.HasChanges() {
		return nil
	}
	fmt.Println(ui.RenderDocsMigrationNeeded(plan))
	return nil
}
