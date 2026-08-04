// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import "testing"

// Scenario: roadmap generate regenerates without committing
func TestRsh_GenerateRegeneratesWithoutCommitting(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	handEdit(t, dir) // a drifted ROADMAP.md, committed — HEAD itself is stale
	commit(t, dir, "hand-edited ROADMAP.md")
	before := rshCommitCount(t, dir)

	rdsGenerate(t, bin, dir)

	if got := rshCommitCount(t, dir); got != before {
		t.Fatalf("generate must never commit, want %d got %d", before, got)
	}
	if rshGitOut(t, dir, "status", "--porcelain", "--", "ROADMAP.md") == "" {
		t.Fatal("ROADMAP.md must be left modified in the working tree")
	}
}

// Scenario: roadmap generate works while a merge is in progress
func TestRsh_GenerateWorksDuringAMerge(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	head := rshGitOut(t, dir, "rev-parse", "HEAD")
	mustWrite(t, dir+"/.git/MERGE_HEAD", head+"\n")

	out, code := runCent(t, bin, dir, "roadmap", "generate")
	if code != 0 {
		t.Fatalf("generate exit=%d\n%s", code, out)
	}
	if rshCommitCount(t, dir) != 1 {
		t.Fatal("generate must never attempt a commit, even mid-merge")
	}
}
