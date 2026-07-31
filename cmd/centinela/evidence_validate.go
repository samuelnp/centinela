package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceValidateCmd = &cobra.Command{
	Use:   "validate <feature>",
	Short: "Validate every evidence file for a feature; non-zero exit on issues",
	Args:  cobra.ExactArgs(1),
	RunE:  runEvidenceValidate,
}

func init() {
	evidenceCmd.AddCommand(evidenceValidateCmd)
}

// runEvidenceValidate mirrors internal/workflow/validate_orchestration.go: it
// loads the project config and passes config.UIPaths(cfg) so the CLI applies
// the same ux-ui-specialist output rule `centinela complete` applies. A config
// load error is RETURNED, never downgraded to nil paths — a silently degraded
// validator is exactly the failure mode this command exists to prevent.
// config.Load already returns defaults for a missing centinela.toml, and
// config.UIPaths falls back to its built-in defaults for an empty ui_paths.
func runEvidenceValidate(_ *cobra.Command, args []string) error {
	feature := args[0]
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	hints := evidence.ValidateFeature(feature, config.UIPaths(cfg))
	if len(hints) == 0 {
		fmt.Fprintf(os.Stdout, "evidence ok for %q\n", feature)
		return nil
	}
	for _, h := range hints {
		fmt.Fprintln(os.Stderr, h.String())
	}
	return fmt.Errorf("evidence validation failed: %d issue(s)", len(hints))
}
