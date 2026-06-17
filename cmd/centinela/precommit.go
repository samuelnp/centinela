package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/gitdiff"
	"github.com/samuelnp/centinela/internal/ui"
)

// precommitCmd runs the gate suite over the staged index and exits non-zero on
// a fail-severity gate, blocking the commit. It is the body of the installed
// .git/hooks/pre-commit. A git failure (not a repo / staged diff failed)
// degrades to a notice + exit 0 — a pre-commit must never false-block.
var precommitCmd = &cobra.Command{
	Use:           "precommit",
	Short:         "Run gates over the staged index, blocking the commit on failure",
	RunE:          runPrecommit,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(precommitCmd)
}

func runPrecommit(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	set, summary, err := gitdiff.Default.ChangedFilesStaged()
	if err != nil {
		return err
	}
	if summary.Degrade != "" {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"centinela precommit: %s — nothing to gate, skipping\n", summary.Degrade)
		return nil
	}
	results := appendAuditGate(precommitCfg(cfg), gates.RunWithFilter(precommitCfg(cfg), set))
	renderPrecommit(cmd, summary, results)
	if !gates.AllPassed(results) {
		return fmt.Errorf("precommit gate failed")
	}
	return nil
}

func renderPrecommit(cmd *cobra.Command, summary gitdiff.Summary, results []gates.Result) {
	out := cmd.ErrOrStderr()
	fmt.Fprintf(out, "centinela precommit — %d staged file(s)\n", summary.Files)
	for _, r := range results {
		fmt.Fprintln(out, ui.RenderGateResult(r))
	}
}
