package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	docsLintChanged bool
	docsLintFull    bool
	docsLintJSON    bool
)

var docsLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Report exported identifiers that carry no doc comment",
	Long: "Runs the docstring gate over the changed-file set — the same check " +
		"centinela validate runs, never a second implementation.\n\n" +
		"--full instead reports the whole-repo backlog. That is a measurement, " +
		"not a gate: it always exits 0.",
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runDocsLint,
}

func init() {
	docsLintCmd.Flags().BoolVar(&docsLintChanged, "changed", false,
		"Inspect only files changed since the diff base (the default)")
	docsLintCmd.Flags().BoolVar(&docsLintFull, "full", false,
		"Report the whole-repo backlog instead of gating; always exits 0")
	docsLintCmd.Flags().BoolVar(&docsLintJSON, "json", false,
		"Emit the report as JSON")
	docsCmd.AddCommand(docsLintCmd)
}

func runDocsLint(_ *cobra.Command, _ []string) error {
	if docsLintChanged && docsLintFull {
		return fmt.Errorf("--changed and --full are mutually exclusive")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if docsLintFull {
		return runDocsLintFull(cfg)
	}
	return runDocsLintChanged(cfg)
}

// runDocsLintChanged runs the real gate over the changed-file scope and exits
// non-zero only on Fail, so the CLI verdict and centinela validate agree.
func runDocsLintChanged(cfg *config.Config) error {
	filter, header := resolveDiffFilter(cfg, config.ModeChanged)
	result, rep := gates.CheckDocstring(cfg, filter)
	if docsLintJSON {
		// JSON goes to stdout and stays valid; the failure reason goes to
		// stderr via main, so the exit code still reflects the verdict.
		if err := emitDocsLintJSON(docsLintPayload("changed", result, rep)); err != nil {
			return err
		}
	} else {
		fmt.Println(ui.StyleBold.Render("Docstring gate " + header))
		fmt.Println(ui.RenderGateResult(result))
		printDocsLintPassDetails(result)
	}
	if result.Status == gates.Fail {
		return fmt.Errorf("docstring gate failed — document the identifiers listed above")
	}
	return nil
}

// printDocsLintPassDetails echoes the detail lines of a non-failing result.
// RenderGateResult only expands details for Fail, which would hide the
// //centinela:nodoc exemption list on a passing run — and an opt-out nobody
// can see is a hole, which is the whole reason the gate reports them.
func printDocsLintPassDetails(r gates.Result) {
	if r.Status == gates.Fail {
		return
	}
	for _, d := range r.Details {
		fmt.Println(ui.StyleMuted.Render("  · " + d))
	}
}
