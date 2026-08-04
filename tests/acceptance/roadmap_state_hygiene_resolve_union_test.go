// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// rshBacklogDoc renders a roadmap.json with one schedulable phase P1 (fixed)
// plus a Backlog phase holding entries, one object per line so two sides that
// both append after the same line collide in git's merge.
func rshBacklogDoc(entries ...string) string {
	return "{\"phases\":[\n{\"name\":\"P1\",\"features\":[{\"name\":\"x\"}]},\n" +
		"{\"name\":\"Backlog\",\"features\":[\n" + strings.Join(entries, ",\n") + "\n]}\n]}\n"
}

func rshFinding(slug, deferredAt string) string {
	return `{"name":"` + slug + `","summary":"s","deferredAt":"` + deferredAt + `"}`
}

// Scenario: resolve unions Backlog findings from both sides of a conflict
func TestRsh_ResolveUnionsBothSidesOfARealConflict(t *testing.T) {
	bin := buildCent(t)
	base := []string{
		rshFinding("b1", "2026-01-01T00:00:00Z"), rshFinding("b2", "2026-01-02T00:00:00Z"),
		rshFinding("b3", "2026-01-03T00:00:00Z"), rshFinding("b4", "2026-01-04T00:00:00Z"),
		rshFinding("b5", "2026-01-05T00:00:00Z"),
	}
	ours := append(append([]string{}, base...),
		rshFinding("o1", "2026-02-01T00:00:00Z"), rshFinding("o2", "2026-02-02T00:00:00Z"))
	theirs := append(append([]string{}, base...),
		rshFinding("t1", "2026-03-01T00:00:00Z"), rshFinding("t2", "2026-03-02T00:00:00Z"),
		rshFinding("t3", "2026-03-03T00:00:00Z"))
	dir := rshConflictRepo(t, rshBacklogDoc(base...), rshBacklogDoc(ours...), rshBacklogDoc(theirs...))
	if !rshUnmerged(t, dir) {
		t.Fatal("fixture bug: roadmap.json must be a real git conflict before resolve runs")
	}

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("resolve exit=%d\n%s", code, out)
	}

	body := mustRead(t, dir+"/.workflow/roadmap.json")
	for _, slug := range []string{"b1", "b2", "b3", "b4", "b5", "o1", "o2", "t1", "t2", "t3"} {
		if !strings.Contains(body, `"`+slug+`"`) {
			t.Fatalf("finding %q was dropped:\n%s", slug, body)
		}
	}
	if strings.Contains(body, "<<<<<<<") {
		t.Fatalf("no conflict markers may survive:\n%s", body)
	}
	if got := rshGitOut(t, dir, "diff", "--cached", "--name-only"); got != ".workflow/roadmap.json\nROADMAP.md" {
		t.Fatalf("both roadmap-state paths must be staged, got %q", got)
	}
	containsAll(t, out, "10", "2 from our side", "3 from theirs")
}
