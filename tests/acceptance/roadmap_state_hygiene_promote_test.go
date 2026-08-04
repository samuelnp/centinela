// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import "testing"

// Scenario: a promote also commits the analysis and quality artifacts it rewrote
func TestRsh_PromoteCommitsAnalysisAndQualityArtifacts(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)

	out, code := runCent(t, bin, dir, "roadmap", "promote", "worth-doing",
		"--phase", "Phase 2", "--scores", "8,8,8,8,8,8")
	if code != 0 {
		t.Fatalf("promote exit=%d\n%s", code, out)
	}

	changed := rshChangedPaths(t, dir, "HEAD")
	containsAll(t, joinLines(changed), ".workflow/roadmap-analysis.json", ".workflow/roadmap-quality.json")
	for _, p := range changed {
		if !rshStatePaths[p] {
			t.Fatalf("commit touched a path outside the declared roadmap-state pathspec: %q (all: %v)", p, changed)
		}
	}
}

func joinLines(paths []string) string {
	out := ""
	for _, p := range paths {
		out += p + "\n"
	}
	return out
}
