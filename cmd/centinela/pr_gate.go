package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/gitdiff"
	"github.com/samuelnp/centinela/internal/ui"
)

// prGateCmd runs the gate suite over the PR's changed-since-base files and
// prints a deterministic Markdown verdict to stdout (for a CI PR comment). It
// exits 1 on a fail-severity gate, or on a warn when [pr_gate] fail_on_warning.
// A diff-resolution failure degrades to a full scan with a notice on stderr.
var prGateCmd = &cobra.Command{
	Use:           "pr-gate",
	Short:         "Render a Markdown gate verdict for a pull request",
	RunE:          runPrGate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(prGateCmd)
}

func runPrGate(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	set, summary, err := gitdiff.Default.ChangedFiles(cfg.Validate.DiffBase, true)
	if err != nil {
		return err
	}
	if summary.Degrade != "" {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"centinela pr-gate: %s — full repo scan\n", summary.Degrade)
		set = nil
	}
	results := appendAuditGate(cfg, gates.RunWithFilter(cfg, set))
	fmt.Fprint(cmd.OutOrStdout(), ui.RenderGatesMarkdown(results))
	if prGateFails(cfg, results) {
		return fmt.Errorf("pr-gate failed")
	}
	return nil
}

// prGateFails reports whether the verdict should produce a non-zero exit: any
// Fail, or any Warn when fail_on_warning is set.
func prGateFails(cfg *config.Config, results []gates.Result) bool {
	for _, r := range results {
		if r.Status == gates.Fail {
			return true
		}
		if cfg.PrGate.FailOnWarning && r.Status == gates.Warn {
			return true
		}
	}
	return false
}
