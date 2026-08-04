// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// rshPhaseDoc renders a roadmap.json with one named schedulable phase whose
// feature description is desc, plus an untouched Backlog phase.
func rshPhaseDoc(phaseName, desc string) string {
	return `{"phases":[{"name":"` + phaseName + `","features":[{"name":"x","description":"` + desc +
		`"}]},{"name":"Backlog","features":[]}]}`
}

// Scenario: resolve refuses when a schedulable phase diverged on both sides
func TestRsh_ResolveRefusesOnARealPhaseConflict(t *testing.T) {
	bin := buildCent(t)
	phase := "Phase 13: Lighter Centinela"
	base := rshPhaseDoc(phase, "base")
	dir := rshConflictRepo(t, base, rshPhaseDoc(phase, "ours"), rshPhaseDoc(phase, "theirs"))
	if !rshUnmerged(t, dir) {
		t.Fatal("fixture bug: not a real conflict")
	}
	before := mustRead(t, dir+"/.workflow/roadmap.json")
	stagedBefore := rshGitOut(t, dir, "diff", "--cached", "--name-only")

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code == 0 {
		t.Fatalf("resolve must refuse a real both-sides phase divergence, got exit 0: %s", out)
	}
	containsAll(t, out, phase)
	if got := mustRead(t, dir+"/.workflow/roadmap.json"); got != before {
		t.Fatalf("roadmap.json must be left byte-identical (markers intact):\ngot:  %q\nwant: %q", got, before)
	}
	if got := rshGitOut(t, dir, "diff", "--cached", "--name-only"); got != stagedBefore {
		t.Fatalf("resolve must not change the staged set, want %q got %q", stagedBefore, got)
	}
}

// Scenario: resolve refuses when one side of the conflict is not valid JSON
func TestRsh_ResolveRefusesAnUnparseableSide(t *testing.T) {
	bin := buildCent(t)
	base := rshPhaseDoc("P1", "base")
	ours := rshPhaseDoc("P1", "ours")
	theirs := "{not valid json"
	dir := rshConflictRepo(t, base, ours, theirs)
	if !rshUnmerged(t, dir) {
		t.Fatal("fixture bug: not a real conflict")
	}
	stagedBefore := rshGitOut(t, dir, "diff", "--cached", "--name-only")

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code == 0 {
		t.Fatalf("resolve must refuse an unparseable side, got exit 0: %s", out)
	}
	if !strings.Contains(strings.ToLower(out), "their") {
		t.Fatalf("message must name the unparseable side:\n%s", out)
	}
	if got := rshGitOut(t, dir, "diff", "--cached", "--name-only"); got != stagedBefore {
		t.Fatalf("resolve must not change the staged set, want %q got %q", stagedBefore, got)
	}
}
