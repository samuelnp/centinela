package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/migration"
	"github.com/samuelnp/centinela/internal/setup"
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
	docsCount, setupCount := 0, 0
	if !exists("PROJECT.md.template") && !exists("PROJECT.md") && !exists("centinela.toml") && !exists(".claude") && !exists("opencode.json") && !exists(".opencode") {
		return nil
	}
	if _, err := os.Stat("PROJECT.md.template"); err == nil {
		plan, err := migration.BuildPlan(".")
		if err == nil {
			docsCount = len(plan.Items)
		}
	}
	sync, err := setup.BuildSyncPlan("both")
	if err == nil {
		setupCount = len(sync.Items)
	}
	if docsCount+setupCount == 0 {
		return nil
	}
	fmt.Println(ui.RenderMigrationNeeded(docsCount, setupCount))
	return nil
}
