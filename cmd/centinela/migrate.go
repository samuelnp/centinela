package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/migration"
	"github.com/samuelnp/centinela/internal/setup"
	"github.com/samuelnp/centinela/internal/ui"
)

var applyFullMigration bool
var fullAgent string

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run project migrations",
	RunE:  runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVar(&applyFullMigration, "apply", false, "Apply changes")
	migrateCmd.Flags().StringVar(&fullAgent, "agent", "both", "Scope for setup: claude, opencode, or both")
}

func runMigrate(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown migrate argument %q", args[0])
	}
	if !isValidAgent(fullAgent) {
		return fmt.Errorf("invalid --agent %q (use: claude|opencode|both)", fullAgent)
	}
	docsPlan, err := migration.BuildPlan(".")
	if err != nil {
		return err
	}
	setupPlan, err := setup.BuildSyncPlan(fullAgent)
	if err != nil {
		return err
	}
	if !docsPlan.HasChanges() && !setupPlan.HasChanges() {
		fmt.Println(ui.RenderSuccess("Managed docs and setup assets are already up to date."))
		return nil
	}
	if docsPlan.HasChanges() {
		fmt.Println(ui.RenderDocsMigrationPlan(docsPlan, applyFullMigration))
	}
	if setupPlan.HasChanges() {
		fmt.Println(ui.RenderSetupMigrationPlan(setupPlan, applyFullMigration))
	}
	if !applyFullMigration {
		return nil
	}
	if docsPlan.HasChanges() {
		if err := migration.Apply(".", docsPlan); err != nil {
			return err
		}
	}
	if setupPlan.HasChanges() {
		if err := setup.ApplySync(setupPlan); err != nil {
			return err
		}
	}
	if cfg, _ := config.Load(); cfg != nil && cfg.Workflow.UseWorktrees {
		if err := syncWorktreeIgnores("."); err != nil {
			return err
		}
	}
	fmt.Println(ui.RenderSuccess("Full migration applied."))
	return nil
}
