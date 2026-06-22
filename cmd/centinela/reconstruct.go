package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/reconstruct"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	reconIn   string
	reconOut  string
	reconJSON bool
)

// reconstructCmd derives a behavioral spec corpus skeleton (one .feature + one
// brief stub per significant surface) from the analyze Inventory, writing it
// into a review dir. It never clobbers hand-authored specs (skip-if-exists) and
// emits no LLM call.
var reconstructCmd = &cobra.Command{
	Use:   "reconstruct",
	Short: "Reconstruct Gherkin spec skeletons and brief stubs from the codebase inventory",
	Long: "Reads the deterministic Inventory written by `centinela analyze` and writes a\n" +
		"behavioral spec corpus skeleton (one specs/<slug>.feature + one\n" +
		"docs/features/<slug>.md per significant surface) into a review dir, with honest\n" +
		"`# TODO: confirm` gaps and no LLM call. Hand-authored specs are never clobbered.",
	RunE:          runReconstruct,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	reconstructCmd.Flags().StringVar(&reconIn, "in", analyze.DefaultOutPath, "Path to the inventory JSON")
	reconstructCmd.Flags().StringVar(&reconOut, "out", reconstruct.DefaultOutRoot, "Review-dir root for the reconstructed corpus")
	reconstructCmd.Flags().BoolVar(&reconJSON, "json", false, "Emit the reconstruction as JSON instead of a summary")
	rootCmd.AddCommand(reconstructCmd)
}

func runReconstruct(cmd *cobra.Command, _ []string) error {
	inv, err := analyze.Load(reconIn)
	if err != nil {
		if errors.Is(err, analyze.ErrNoInventory) {
			return fmt.Errorf("%w — run `centinela analyze` first", err)
		}
		return err
	}
	r := reconstruct.NewReconstructor().Reconstruct(inv)
	written, skipped, err := reconstruct.WriteCorpus(reconOut, r)
	if err != nil {
		return fmt.Errorf("reconstruct: cannot write corpus: %w", err)
	}
	r.Written, r.Skipped = written, skipped
	out := cmd.OutOrStdout()
	if reconJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	}
	fmt.Fprintln(out, ui.RenderReconstructionSummary(r))
	return nil
}
