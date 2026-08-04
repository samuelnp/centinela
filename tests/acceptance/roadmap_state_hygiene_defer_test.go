// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"testing"
)

// Scenario: a deferral regenerates ROADMAP.md and commits roadmap state by itself
func TestRsh_DeferRegeneratesAndCommits(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	mustWrite(t, dir+"/centinela.toml", rdsToml("fail"))

	out, code := runCent(t, bin, dir, "roadmap", "defer", "flaky-thing", "--summary", "one line")
	if code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}

	// The roadmap_drift gate byte-compares ROADMAP.md against the same
	// renderer defer just ran, without invoking `roadmap generate` again.
	vout, vcode := rdsValidate(t, bin, dir)
	containsAll(t, vout, "roadmap_drift", "in sync")
	if vcode != 0 {
		t.Fatalf("post-defer validate must exit 0\n%s", vout)
	}

	if got := rshCommitCount(t, dir); got != 2 {
		t.Fatalf("want exactly one new commit (2 total), got %d", got)
	}
	changed := rshChangedPaths(t, dir, "HEAD")
	if len(changed) != 2 || changed[0] != ".workflow/roadmap.json" || changed[1] != "ROADMAP.md" {
		t.Fatalf("commit changed paths = %v, want exactly the roadmap-state pair", changed)
	}
	if got := rshLastMsg(t, dir); got != "chore(roadmap): defer flaky-thing" {
		t.Fatalf("message = %q", got)
	}
	if !rshClean(t, dir) {
		t.Fatal("roadmap state must be clean after the mutation")
	}
}
