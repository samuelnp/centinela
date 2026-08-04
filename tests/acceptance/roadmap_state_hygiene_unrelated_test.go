// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import "testing"

// Scenario: unrelated staged and unstaged changes survive the mutation commit
func TestRsh_UnrelatedStagedAndUnstagedChangesSurvive(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)

	mustWrite(t, dir+"/internal_example_thing.go", "package example\n")
	runGit(t, dir, "add", "internal_example_thing.go")
	mustWrite(t, dir+"/README.md", "unstaged edit\n")

	out, code := runCent(t, bin, dir, "roadmap", "defer", "another-thing", "--summary", "x")
	if code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}

	changed := rshChangedPaths(t, dir, "HEAD")
	for _, p := range changed {
		if p == "internal_example_thing.go" || p == "README.md" {
			t.Fatalf("mutation commit must not touch unrelated paths: %v", changed)
		}
	}
	if got := rshGitOut(t, dir, "diff", "--cached", "--name-only"); got != "internal_example_thing.go" {
		t.Fatalf("unrelated staged file must still be staged, got %q", got)
	}
	if got := rshGitOut(t, dir, "status", "--porcelain", "--", "README.md"); got == "" {
		t.Fatal("README.md must still show as modified and unstaged")
	}
}
