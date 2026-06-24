package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	adoptForce bool
	adoptJSON  bool
)

// adoptCmd is the one-time brownfield adoption command. It records the current
// full-repo gate violations as the accepted audit baseline, refusing to overwrite
// an established baseline unless --force is given. The skip-if-exists decision
// lives in audit.Adopt (G7); this cmd only maps the Outcome to exit code + render.
var adoptCmd = &cobra.Command{
	Use:   "adopt",
	Short: "Record today's gate violations as the accepted baseline (one-time brownfield adoption)",
	Long: "Snapshots the current full-repo gate violations as the accepted audit\n" +
		"baseline (.workflow/audit-baseline.json). Refuses to overwrite an existing\n" +
		"baseline unless --force. Run it between `centinela roadmap brownfield` and\n" +
		"your first `centinela start` so day-one validate is not drowned by debt.",
	RunE:          runAdopt,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	adoptCmd.Flags().BoolVar(&adoptForce, "force", false, "Overwrite an existing baseline")
	adoptCmd.Flags().BoolVar(&adoptJSON, "json", false, "Emit the adoption verdict as JSON")
	rootCmd.AddCommand(adoptCmd)
}

func runAdopt(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	o, err := audit.Adopt(cfg, adoptForce)
	if err != nil {
		return err
	}
	if adoptJSON {
		return printAdoptJSON(cmd, o)
	}
	if o.Skipped {
		return fmt.Errorf("baseline already exists at %s — use --force to overwrite", o.Path)
	}
	fmt.Fprintln(cmd.OutOrStdout(), ui.RenderAdoption(o))
	return nil
}
