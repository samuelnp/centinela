package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/synthesize"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	synthIn   string
	synthOut  string
	synthJSON bool
)

// synthesizeCmd infers the best-fit architecture archetype from the analyze
// Inventory and writes a draft PROJECT.md for the user to confirm/correct. It
// never overwrites an existing PROJECT.md (it writes PROJECT.draft.md instead).
var synthesizeCmd = &cobra.Command{
	Use:   "synthesize",
	Short: "Infer the archetype and draft a PROJECT.md from the codebase inventory",
	Long: "Reads the deterministic Inventory written by `centinela analyze`, infers the\n" +
		"best-fit architecture archetype (no LLM), and writes a draft PROJECT.md to\n" +
		"confirm or correct. An existing PROJECT.md is never overwritten.",
	RunE:          runSynthesize,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	synthesizeCmd.Flags().StringVar(&synthIn, "in", analyze.DefaultOutPath, "Path to the inventory JSON")
	synthesizeCmd.Flags().StringVar(&synthOut, "out", synthesize.DefaultTarget, "Target project-definition path")
	synthesizeCmd.Flags().BoolVar(&synthJSON, "json", false, "Emit the inference as JSON instead of writing a draft")
	rootCmd.AddCommand(synthesizeCmd)
}

func runSynthesize(cmd *cobra.Command, _ []string) error {
	inv, err := analyze.Load(synthIn)
	if err != nil {
		if errors.Is(err, analyze.ErrNoInventory) {
			return fmt.Errorf("%w — run `centinela analyze` first", err)
		}
		return err
	}
	inf := synthesize.NewInferer().Infer(inv)
	out := cmd.OutOrStdout()
	if synthJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(inf)
	}
	written, clobbered, err := synthesize.WriteDraft(synthOut, synthesize.Draft(inv, inf))
	if err != nil {
		return fmt.Errorf("synthesize: cannot write draft: %w", err)
	}
	fmt.Fprintln(out, ui.RenderInferenceSummary(inf))
	fmt.Fprintln(out, "wrote "+written)
	if clobbered {
		fmt.Fprintln(out, "note: existing "+synthOut+" preserved; draft written instead")
	}
	return nil
}
