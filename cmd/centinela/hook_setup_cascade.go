package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

// setupRung is one greenfield-cascade rung whose weight the profile scopes:
// satisfied reports whether its artifact is present, and directive/panel are the
// exact lines the strict cascade prints when it is not.
type setupRung struct {
	label     string
	satisfied func() bool
	directive string
	panel     func() string
}

// setupRequiresGrading reports whether the cascade's grading rungs HALT the
// setup flow. Strict halts; guided and outcome advise. The shared resolver
// supplies the config, so a centinela.toml that cannot be parsed halts too —
// when the profile is unknowable the fail-safe direction is more scaffolding.
func setupRequiresGrading() bool {
	cfg, _ := config.LoadForProfile()
	return config.ProfileDefaults(config.ProjectDefaultProfile(cfg)).RequireRoadmapGrading
}

// preRoadmapRungs are the graded rungs that run BEFORE .workflow/roadmap.json is
// read. ROADMAP.md is generated from that json, so under guided its absence is
// advice; under strict the directive and its ordering are unchanged.
func preRoadmapRungs() []setupRung {
	return []setupRung{{
		label:     "ROADMAP.md (run `centinela roadmap generate`)",
		satisfied: func() bool { return exists("ROADMAP.md") },
		directive: "CENTINELA DIRECTIVE: roadmap required. Define roadmap before feature work.",
		panel:     ui.RenderRoadmapNeeded,
	}}
}

// postRoadmapRungs are the graded rungs that run after the roadmap json parses,
// in the same order the pre-profile cascade used.
func postRoadmapRungs() []setupRung {
	return []setupRung{{
		label: ".workflow/roadmap-analysis.{md,json}",
		satisfied: func() bool {
			return evaluatedArtifact(".workflow/roadmap-analysis.md", roadmap.RoadmapAnalysisFile)
		},
		directive: "CENTINELA DIRECTIVE: roadmap analysis required. Delegate to senior product manager.",
		panel:     ui.RenderRoadmapAnalysisNeeded,
	}, {
		label: ".workflow/roadmap-quality.{md,json}",
		satisfied: func() bool {
			return evaluatedArtifact(".workflow/roadmap-quality.md", roadmap.RoadmapQualityFile)
		},
		directive: "CENTINELA DIRECTIVE: roadmap quality required. Delegate to roadmap quality evaluator.",
		panel:     ui.RenderRoadmapQualityNeeded,
	}, {
		label:     "docs/architecture/production-readiness-prompt.md",
		satisfied: func() bool { return exists("docs/architecture/production-readiness-prompt.md") },
		directive: "CENTINELA DIRECTIVE: configure production-readiness prompt before continuing.",
		panel:     ui.RenderProductionReadinessSetupNeeded,
	}}
}

// runSetupRungs walks rungs in order. When graded, the first unmet rung prints
// its directive and panel and halts, byte-identically to the pre-profile
// cascade. Otherwise every unmet rung's label is collected for one consolidated
// advisory and the flow continues to the roadmap checkpoint.
func runSetupRungs(rungs []setupRung, graded bool) (halt bool, missing []string) {
	for _, rung := range rungs {
		if rung.satisfied() {
			continue
		}
		if graded {
			fmt.Println(rung.directive)
			fmt.Println(rung.panel())
			return true, nil
		}
		missing = append(missing, rung.label)
	}
	return false, missing
}

// emitSetupAdvisory prints one line naming every optional setup artifact a
// guided/outcome project lacks. Advisory only — nothing is blocked.
func emitSetupAdvisory(missing []string) {
	if len(missing) == 0 {
		return
	}
	fmt.Println(ui.StyleMuted.Render(
		"Advisory: optional setup artifacts missing — " + strings.Join(missing, ", ")))
}
