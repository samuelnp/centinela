package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/setup"
	"github.com/samuelnp/centinela/internal/ui"
)

var applySetupMigration bool
var setupAgent string

var migrateSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Migrate managed setup assets (hooks/config/plugin/agents)",
	RunE:  runMigrateSetup,
}

func init() {
	migrateCmd.AddCommand(migrateSetupCmd)
	migrateSetupCmd.Flags().BoolVar(&applySetupMigration, "apply", false, "Apply changes")
	migrateSetupCmd.Flags().StringVar(&setupAgent, "agent", "both", "Scope: claude, opencode, or both")
}

func runMigrateSetup(_ *cobra.Command, _ []string) error {
	if !isValidAgent(setupAgent) {
		return fmt.Errorf("invalid --agent %q (use: claude|opencode|both)", setupAgent)
	}
	plan, err := setup.BuildSyncPlan(setupAgent)
	if err != nil {
		return err
	}
	if !plan.HasChanges() {
		fmt.Println(ui.RenderSuccess("Managed setup assets are already up to date."))
		return nil
	}
	fmt.Println(ui.RenderSetupMigrationPlan(plan, applySetupMigration))
	if !applySetupMigration {
		return nil
	}
	if err := setup.ApplySync(plan); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess("Setup migration applied."))
	return nil
}
