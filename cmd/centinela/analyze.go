package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/ui"
)

var analyzeOut string

// analyzeCmd performs a deterministic, read-only scan of the current repo and
// writes a schemaVersion-tagged Inventory to the well-known output path, then
// prints a concise summary. It is diagnostic: sub-detector failures degrade
// best-effort and still exit 0. Only an unreadable root or an un-writable
// output path is a hard error (non-zero exit, no partial artifact).
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Scan the repo and write a deterministic codebase inventory",
	Long: "Walks the current directory (read-only) and writes a machine-readable\n" +
		"Inventory (languages, manifests, locales, layout, dependency graph) to\n" +
		".workflow/analysis.json. No LLM call; output is byte-stable across re-runs.",
	RunE:          runAnalyze,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	analyzeCmd.Flags().StringVar(&analyzeOut, "out", analyze.DefaultOutPath,
		"Path to write the inventory JSON")
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, _ []string) error {
	inv, err := analyze.Analyze(".")
	if err != nil {
		return fmt.Errorf("analyze: unreadable root: %w", err)
	}
	if err := analyze.Save(analyzeOut, inv); err != nil {
		return fmt.Errorf("analyze: cannot write %s: %w", analyzeOut, err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), ui.RenderInventorySummary(inv))
	fmt.Fprintln(cmd.OutOrStdout(), "wrote "+analyzeOut)
	return nil
}
