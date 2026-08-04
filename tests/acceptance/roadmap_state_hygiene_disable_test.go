// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os"
	"testing"
)

// Scenario: disable_auto_commit skips the commit but never skips regeneration
func TestRsh_DisableAutoCommitSkipsCommitNotRegeneration(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	mustWrite(t, dir+"/centinela.toml", "[workflow]\ndisable_auto_commit = true\n")
	before := rshCommitCount(t, dir)

	out, code := runCent(t, bin, dir, "roadmap", "defer", "policy-thing", "--summary", "x")
	if code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}
	if got := rshCommitCount(t, dir); got != before {
		t.Fatalf("no new commit may be created, want %d got %d", before, got)
	}
	if _, err := os.Stat(dir + "/ROADMAP.md"); err != nil {
		t.Fatalf("ROADMAP.md must still be regenerated: %v", err)
	}
	containsAll(t, out, "left uncommitted")
}
