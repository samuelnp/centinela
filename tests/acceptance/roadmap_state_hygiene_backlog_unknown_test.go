// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import "testing"

// Scenario: a finding with no deferredAt is reported as unknown age and counted stale
func TestRsh_BacklogUnknownDeferredAtSortsFirstAndCountsStale(t *testing.T) {
	bin := buildCent(t)
	// finding-a: no deferredAt field. finding-b: unparseable deferredAt.
	// finding-c: a real, recent (non-stale) date — must NOT appear as a row.
	dir := rshBacklogFixture(t, -1, -2, 2)
	out, code := runCent(t, bin, dir, "roadmap", "backlog", "--stale")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	containsAll(t, out, "unknown", "finding-a", "finding-b")
	// Only the two unknown-age findings are ROWS under --stale; finding-c (2
	// days, non-stale) may still be named by the footer as the oldest KNOWN
	// finding, so check the row list (everything before the blank separator)
	// rather than the whole output.
	rows := out
	if i := indexOf(out, "\n\n"); i >= 0 {
		rows = out[:i]
	}
	if contains(rows, "finding-c") {
		t.Fatalf("the non-stale finding must not appear as a row under --stale:\n%s", out)
	}
}
