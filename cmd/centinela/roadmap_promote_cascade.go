package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

// promoteRequiresGrading reports whether `roadmap promote` must find the roadmap
// grading artifacts already on disk. Strict must; guided and outcome seed them
// instead. The shared resolver supplies the config, so an unparseable
// centinela.toml requires grading — the same fail-closed direction every other
// profile surface takes.
func promoteRequiresGrading() bool {
	cfg, _ := config.LoadForProfile()
	return config.ProfileDefaults(config.ProjectDefaultProfile(cfg)).RequireRoadmapGrading
}

// reportGradingAfterPromote re-validates the artifacts promote just wrote.
//
// Under strict this is the pre-existing post-write check and its error text is
// unchanged. Under guided/outcome the artifacts are advisory and cover only the
// features actually promoted, so a coverage complaint here would refuse a
// promotion that already succeeded on disk — it is reported as advice instead.
func reportGradingAfterPromote(r *roadmap.Roadmap) error {
	if promoteRequiresGrading() {
		if err := roadmap.ValidateAnalysis(r); err != nil {
			return fmt.Errorf("promote wrote files but validate failed: %w", err)
		}
		if err := roadmap.ValidateQuality(r); err != nil {
			return fmt.Errorf("promote wrote files but validate failed: %w", err)
		}
		return nil
	}
	if missing := missingGradingArtifacts(r); len(missing) > 0 {
		fmt.Println(ui.StyleMuted.Render(
			"Advisory: the roadmap grading artifacts are incomplete — optional under " +
				"this profile; run `centinela roadmap validate` for detail."))
	}
	return nil
}
