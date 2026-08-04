// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: resolve keeps one entry when both sides added the same slug
func TestRsh_ResolveKeepsOneEntryForASharedSlug(t *testing.T) {
	bin := buildCent(t)
	base := []string{rshFinding("anchor", "2026-01-01T00:00:00Z")}
	ours := append(append([]string{}, base...), rshFinding("same-thing", "2026-02-02T00:00:00Z"))
	theirs := append(append([]string{}, base...), rshFinding("same-thing", "2026-02-01T00:00:00Z"))
	dir := rshConflictRepo(t, rshBacklogDoc(base...), rshBacklogDoc(ours...), rshBacklogDoc(theirs...))
	if !rshUnmerged(t, dir) {
		t.Fatal("fixture bug: not a real conflict")
	}

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("resolve exit=%d\n%s", code, out)
	}
	body := mustRead(t, dir+"/.workflow/roadmap.json")
	if strings.Count(body, `"same-thing"`) != 1 {
		t.Fatalf("want exactly one entry for the shared slug:\n%s", body)
	}
	if !strings.Contains(body, `"same-thing","summary":"s","deferredAt":"2026-02-01T00:00:00Z"`) {
		t.Fatalf("the earlier deferredAt must survive:\n%s", body)
	}
}

// Scenario: resolve honours a one-sided deletion
func TestRsh_ResolveHonoursAOneSidedDeletion(t *testing.T) {
	bin := buildCent(t)
	base := []string{rshFinding("promoted-away", "2026-01-01T00:00:00Z"), rshFinding("anchor", "2026-01-02T00:00:00Z")}
	// ours deletes promoted-away and adds a new finding at the same insertion
	// point theirs also touches, so the file registers a real git conflict.
	ours := []string{rshFinding("anchor", "2026-01-02T00:00:00Z"), rshFinding("ours-new", "2026-02-01T00:00:00Z")}
	theirs := append(append([]string{}, base...), rshFinding("theirs-new", "2026-03-01T00:00:00Z"))
	dir := rshConflictRepo(t, rshBacklogDoc(base...), rshBacklogDoc(ours...), rshBacklogDoc(theirs...))
	if !rshUnmerged(t, dir) {
		t.Fatal("fixture bug: not a real conflict")
	}

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("resolve exit=%d\n%s", code, out)
	}
	body := mustRead(t, dir+"/.workflow/roadmap.json")
	if strings.Contains(body, "promoted-away") {
		t.Fatalf("a slug deleted on exactly one side must stay deleted:\n%s", body)
	}
	containsAll(t, body, "anchor", "ours-new", "theirs-new")
}
