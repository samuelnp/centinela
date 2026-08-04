package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

// roadmapGradingGuard applies the greenfield cascade's GRADING rungs — the
// senior-PM roadmap analysis and the quality evaluation — at the weight the
// resolved profile asks for.
//
// Under strict (RequireRoadmapGrading) both refusals are byte-identical to the
// pre-profile cascade. Under guided/outcome they become one advisory line and
// `start` proceeds: the artifacts describe the work rather than verify it, so
// demanding them before the first line of code is process, not proof. Every
// refusal that reads the roadmap itself — Backlog, draft, bootstrap phase,
// bootstrap completeness, unmet dependencies — is unconditional and lives in
// workflowOrderForFeature, untouched by this function.
func roadmapGradingGuard(r *roadmap.Roadmap, profile string) error {
	if config.ProfileDefaults(profile).RequireRoadmapGrading {
		if err := roadmap.ValidateAnalysis(r); err != nil {
			return fmt.Errorf("greenfield project requires roadmap senior PM analysis: %w", err)
		}
		if err := roadmap.ValidateQuality(r); err != nil {
			return fmt.Errorf("greenfield project requires roadmap quality evaluation: %w", err)
		}
		return nil
	}
	if missing := missingGradingArtifacts(r); len(missing) > 0 {
		fmt.Println(ui.StyleMuted.Render(
			"Advisory: no " + strings.Join(missing, " and ") + " — optional under the " +
				profile + " profile; run `centinela roadmap validate` for detail."))
	}
	return nil
}

// missingGradingArtifacts names the grading artifacts that would have blocked
// under strict, in cascade order, for the advisory line.
func missingGradingArtifacts(r *roadmap.Roadmap) []string {
	var missing []string
	if err := roadmap.ValidateAnalysis(r); err != nil {
		missing = append(missing, "roadmap senior PM analysis")
	}
	if err := roadmap.ValidateQuality(r); err != nil {
		missing = append(missing, "roadmap quality evaluation")
	}
	return missing
}
