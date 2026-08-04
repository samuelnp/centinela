// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"testing"
	"time"
)

// rshBacklogFixture builds a roadmap with Backlog findings deferred the given
// number of days ago. -1 means "no deferredAt field"; -2 means an unparseable
// deferredAt string — both must read as unknown age.
func rshBacklogFixture(t *testing.T, agesDays ...int) string {
	t.Helper()
	entries := ""
	for i, d := range agesDays {
		if i > 0 {
			entries += ","
		}
		name := "finding-" + string(rune('a'+i))
		switch {
		case d == -1:
			entries += `{"name":"` + name + `","summary":"s"}`
		case d == -2:
			entries += `{"name":"` + name + `","summary":"s","deferredAt":"not-a-date"}`
		default:
			when := time.Now().AddDate(0, 0, -d).Format(time.RFC3339)
			entries += `{"name":"` + name + `","summary":"s","deferredAt":"` + when + `"}`
		}
	}
	rm := `{"phases":[{"name":"Backlog","features":[` + entries + `]}]}`
	return acceptanceDir(t, rm)
}

// Scenario: roadmap backlog lists deferred findings oldest first with their age
func TestRsh_BacklogListsOldestFirstWithAge(t *testing.T) {
	bin := buildCent(t)
	dir := rshBacklogFixture(t, 40, 20, 2)
	out, code := runCent(t, bin, dir, "roadmap", "backlog")
	if code != 0 {
		t.Fatalf("backlog exit=%d\n%s", code, out)
	}
	oldest := indexOf(out, "finding-a")
	middle := indexOf(out, "finding-b")
	newest := indexOf(out, "finding-c")
	if !(oldest < middle && middle < newest) {
		t.Fatalf("findings must list oldest first:\n%s", out)
	}
	containsAll(t, out, "40d", "20d", "2d", "3 findings", "2 older than 14d")
}

// Scenario: --stale shows only findings older than the default threshold
func TestRsh_BacklogStaleFiltersByDefaultThreshold(t *testing.T) {
	bin := buildCent(t)
	dir := rshBacklogFixture(t, 40, 20, 2)
	out, code := runCent(t, bin, dir, "roadmap", "backlog", "--stale")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	containsAll(t, out, "finding-a", "finding-b")
	if contains(out, "finding-c") {
		t.Fatalf("a 2-day-old finding must not show under --stale:\n%s", out)
	}
}

// Scenario: --older-than overrides the default threshold
func TestRsh_BacklogOlderThanOverridesThreshold(t *testing.T) {
	bin := buildCent(t)
	dir := rshBacklogFixture(t, 40, 20, 2)
	out, code := runCent(t, bin, dir, "roadmap", "backlog", "--stale", "--older-than", "30")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	containsAll(t, out, "finding-a")
	if contains(out, "finding-b") || contains(out, "finding-c") {
		t.Fatalf("only the 40-day finding should show at --older-than 30:\n%s", out)
	}
}
