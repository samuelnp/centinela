package gates

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// roadmapMarkdownFile is the on-disk human-readable roadmap the gate compares
// against generator output. The gate is whole-file, so it ignores the filter.
const roadmapMarkdownFile = "ROADMAP.md"

const roadmapRemediation = "run `centinela roadmap generate`"

// checkRoadmapDrift regenerates ROADMAP.md in memory from roadmap.json and
// byte-compares it with the on-disk file. A load error, a missing file, or any
// byte mismatch is drift; severity maps drift to Fail or Warn.
func checkRoadmapDrift(cfg *config.Config, _ *gitdiff.Set) Result {
	r := Result{Name: "roadmap_drift"}
	rm, err := roadmap.Load()
	if err != nil {
		r.Status = Fail
		r.Message = "roadmap_drift: cannot load " + roadmap.RoadmapFile + ": " + err.Error()
		return r
	}
	want := roadmap.RenderMarkdown(rm)
	got, err := os.ReadFile(roadmapMarkdownFile)
	if err != nil {
		if os.IsNotExist(err) {
			return driftResult(cfg.Gates.RoadmapDrift.Severity,
				roadmapMarkdownFile+" is missing — "+roadmapRemediation+".")
		}
		r.Status = Fail
		r.Message = "roadmap_drift: cannot read " + roadmapMarkdownFile + ": " + err.Error()
		return r
	}
	if bytes.Equal(want, got) {
		r.Status = Pass
		r.Message = roadmapMarkdownFile + " is in sync."
		return r
	}
	line := firstDifferingLine(want, got)
	return driftResult(cfg.Gates.RoadmapDrift.Severity,
		fmt.Sprintf("%s drifted at line %d — %s.", roadmapMarkdownFile, line, roadmapRemediation))
}

// driftResult maps a severity to a Fail/Warn Result carrying msg on both the
// Message (so it renders in warn mode) and Details (for fail-mode listing).
func driftResult(severity, msg string) Result {
	r := Result{Name: "roadmap_drift", Message: msg, Details: []string{msg}}
	if severity == "warn" {
		r.Status = Warn
	} else {
		r.Status = Fail
	}
	return r
}

// firstDifferingLine returns the 1-based line number of the first line that
// differs between want and got (a missing trailing line counts as differing).
func firstDifferingLine(want, got []byte) int {
	w := strings.Split(string(want), "\n")
	g := strings.Split(string(got), "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return i + 1
		}
	}
	return 0
}
