// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import "testing"

// Scenario: a mutation that changes nothing on disk creates no commit
func TestRsh_NoopMutationCreatesNoCommit(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	// Settle ROADMAP.md into HEAD first — otherwise even a no-op mutation would
	// regenerate it from nothing and legitimately commit a first-time write.
	rdsGenerate(t, bin, dir)
	commit(t, dir, "generate ROADMAP.md")
	before := rshCommitCount(t, dir)

	// A no-flag edit is a genuine no-op: roadmap.Edit returns before writing.
	out, code := runCent(t, bin, dir, "roadmap", "edit", "feature-a")
	if code != 0 {
		t.Fatalf("no-op edit exit=%d\n%s", code, out)
	}
	if got := rshCommitCount(t, dir); got != before {
		t.Fatalf("a byte-identical mutation must create no commit, want %d got %d", before, got)
	}
}
