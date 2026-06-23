package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/brownmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	brownIn    string
	brownOut   string
	brownJSON  bool
	brownGoals []string
)

// roadmapBrownfieldCmd emits a DRAFT roadmap for a brownfield repo from the
// analyze Inventory: a Baseline phase of already-built surfaces plus net-new gap
// phase(s) from reconstruct TODO targets and --goal strings. It never clobbers
// the canonical roadmap.json and makes no LLM call.
var roadmapBrownfieldCmd = &cobra.Command{
	Use:   "brownfield",
	Short: "Generate a draft roadmap from the codebase inventory (Baseline + gap phases)",
	Long: "Reads the deterministic Inventory written by `centinela analyze` and writes a\n" +
		"DRAFT roadmap partitioning capability into a schedule-exempt Baseline phase\n" +
		"(already-built surfaces, never re-planned) plus net-new gap phase(s) seeded from\n" +
		"reconstruct `# TODO: confirm` targets and repeatable --goal strings. The canonical\n" +
		".workflow/roadmap.json is never clobbered and no LLM is called.",
	RunE:          runRoadmapBrownfield,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	roadmapBrownfieldCmd.Flags().StringVar(&brownIn, "in", analyze.DefaultOutPath, "Path to the inventory JSON")
	roadmapBrownfieldCmd.Flags().StringVar(&brownOut, "out", brownmap.DefaultDraftPath, "Draft roadmap output path")
	roadmapBrownfieldCmd.Flags().BoolVar(&brownJSON, "json", false, "Emit the draft plan as JSON instead of a summary")
	roadmapBrownfieldCmd.Flags().StringArrayVar(&brownGoals, "goal", nil, "Add a net-new gap feature from a goal (repeatable)")
	roadmapCmd.AddCommand(roadmapBrownfieldCmd)
}

func runRoadmapBrownfield(cmd *cobra.Command, _ []string) error {
	inv, err := analyze.Load(brownIn)
	if err != nil {
		if errors.Is(err, analyze.ErrNoInventory) {
			return fmt.Errorf("%w — run `centinela analyze` first", err)
		}
		return err
	}
	plan := brownmap.NewBrownfielder().Generate(inv, brownGoals)
	if _, err = brownmap.WriteDraft(brownOut, plan); err != nil {
		return fmt.Errorf("roadmap brownfield: cannot write draft: %w", err)
	}
	plan.DraftPath = brownOut
	out := cmd.OutOrStdout()
	if brownJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(plan)
	}
	fmt.Fprintln(out, ui.RenderBrownfieldSummary(plan))
	return nil
}
