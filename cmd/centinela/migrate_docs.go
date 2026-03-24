package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/migration"
	"github.com/samuelnp/centinela/internal/ui"
)

var applyDocsMigration bool

var migrateDocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Migrate managed docs to latest templates",
	RunE:  runMigrateDocs,
}

func init() {
	migrateCmd.AddCommand(migrateDocsCmd)
	migrateDocsCmd.Flags().BoolVar(&applyDocsMigration, "apply", false, "Apply changes")
}

func runMigrateDocs(_ *cobra.Command, _ []string) error {
	plan, err := migration.BuildPlan(".")
	if err != nil {
		return err
	}
	if !plan.HasChanges() {
		fmt.Println(ui.RenderSuccess("Managed docs are already up to date."))
		return nil
	}
	fmt.Println(ui.RenderDocsMigrationPlan(plan, applyDocsMigration))
	if !applyDocsMigration {
		return nil
	}
	if err := migration.Apply(".", plan); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess("Migration applied."))
	return nil
}
