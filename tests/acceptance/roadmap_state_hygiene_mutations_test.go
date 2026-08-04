// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os"
	"strings"
	"testing"
)

// rshStatePaths is every path the feature defines as "roadmap state" (the
// always-present pair plus what promote also rewrites).
var rshStatePaths = map[string]bool{
	".workflow/roadmap.json":          true,
	"ROADMAP.md":                      true,
	".workflow/roadmap-analysis.json": true,
	".workflow/roadmap-quality.json":  true,
	".workflow/roadmap-analysis.md":   true,
	".workflow/roadmap-quality.md":    true,
}

// Scenario: every roadmap.json mutation regenerates and commits
func TestRsh_EveryMutationRegeneratesAndCommits(t *testing.T) {
	bin := buildCent(t)
	cases := []struct {
		verb, subject string
		args          []string
	}{
		{"add", "new-thing", []string{"roadmap", "add", "new-thing", "--phase", "Phase 1"}},
		{"remove", "feature-c", []string{"roadmap", "remove", "feature-c"}},
		{"edit", "feature-a", []string{"roadmap", "edit", "feature-a", "--description", "Updated."}},
		{"move", "feature-b", []string{"roadmap", "move", "feature-b", "--to-phase", "Phase 2"}},
		{"reorder", "feature-b", []string{"roadmap", "reorder", "feature-b", "--before", "feature-a"}},
		{"promote", "worth-doing", []string{"roadmap", "promote", "worth-doing", "--phase", "Phase 2",
			"--scores", "8,8,8,8,8,8"}},
		{"phase add", "Phase 3", []string{"roadmap", "phase", "add", "Phase 3"}},
		{"phase rename", "Phase Two", []string{"roadmap", "phase", "rename", "Phase 2", "Phase Two"}},
		{"phase remove", "Empty Phase", []string{"roadmap", "phase", "remove", "Empty Phase"}},
	}
	for _, tc := range cases {
		t.Run(tc.verb, func(t *testing.T) {
			dir := rshRepo(t, rshBaseRoadmap)
			out, code := runCent(t, bin, dir, tc.args...)
			if code != 0 {
				t.Fatalf("%v exit=%d\n%s", tc.args, code, out)
			}
			if _, err := os.Stat(dir + "/ROADMAP.md"); err != nil {
				t.Fatalf("ROADMAP.md must be regenerated: %v", err)
			}
			if got := rshCommitCount(t, dir); got != 2 {
				t.Fatalf("want exactly one new commit (2 total), got %d", got)
			}
			for _, p := range rshChangedPaths(t, dir, "HEAD") {
				if !rshStatePaths[p] {
					t.Fatalf("commit touched non-roadmap-state path %q", p)
				}
			}
			wantPrefix := "chore(roadmap): " + tc.verb
			if got := rshLastMsg(t, dir); !strings.HasPrefix(got, wantPrefix) {
				t.Fatalf("message = %q, want prefix %q", got, wantPrefix)
			}
		})
	}
}
