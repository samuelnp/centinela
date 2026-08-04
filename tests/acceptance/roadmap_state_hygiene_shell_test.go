// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// shellDoc builds a roadmap whose Backlog phase carries a `note`, plus the
// given extra finding, so the two sides really conflict on the same hunk.
func shellDoc(note, extra string) string {
	entries := `{"name":"b1","summary":"s","deferredAt":"2026-01-01T00:00:00Z"}`
	if extra != "" {
		entries += "," + extra
	}
	return `{"phases":[{"name":"Phase 1","features":[{"name":"alpha"}]},
 {"name":"Backlog","note":"` + note + `","features":[` + entries + `]}]}`
}

const shellOursNew = `{"name":"ours-new","summary":"o","deferredAt":"2026-02-01T00:00:00Z"}`
const shellTheirsNew = `{"name":"theirs-new","summary":"t","deferredAt":"2026-03-01T00:00:00Z"}`

// Scenario: resolve keeps a one-sided edit to the Backlog phase note
//
// The phase SHELL was merged ours-wins rather than three-way, so an incoming
// edit to `note` — a first-class Phase field rendered as the ROADMAP.md phase
// blockquote — was dropped with exit 0 behind a green "✓ Resolved".
func TestRsh_ResolveKeepsIncomingShellEdit(t *testing.T) {
	bin := buildCent(t)
	dir := rshConflictRepo(t,
		shellDoc("ORIGINAL-NOTE", ""),
		shellDoc("ORIGINAL-NOTE", shellOursNew),
		shellDoc("THEIRS-EDITED-NOTE", shellTheirsNew))

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("a one-sided shell edit is not a conflict: exit %d\n%s", code, out)
	}
	body := mustRead(t, dir+"/.workflow/roadmap.json")
	if !strings.Contains(body, "THEIRS-EDITED-NOTE") {
		t.Fatalf("their note edit was silently discarded:\n%s\n%s", out, body)
	}
	if strings.Contains(body, "ORIGINAL-NOTE") {
		t.Fatalf("the base note must not survive an incoming edit:\n%s", body)
	}
	for _, slug := range []string{"b1", "ours-new", "theirs-new"} {
		if !strings.Contains(body, slug) {
			t.Fatalf("%q was dropped:\n%s", slug, body)
		}
	}
	// The note is generated prose: ROADMAP.md must carry it after regeneration.
	if !strings.Contains(mustRead(t, dir+"/ROADMAP.md"), "THEIRS-EDITED-NOTE") {
		t.Fatal("the merged note must reach the regenerated ROADMAP.md")
	}
}

// Scenario: resolve keeps a one-sided local edit to the Backlog phase note
func TestRsh_ResolveKeepsLocalShellEdit(t *testing.T) {
	bin := buildCent(t)
	dir := rshConflictRepo(t,
		shellDoc("ORIGINAL-NOTE", ""),
		shellDoc("OURS-EDITED-NOTE", shellOursNew),
		shellDoc("ORIGINAL-NOTE", shellTheirsNew))

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	body := mustRead(t, dir+"/.workflow/roadmap.json")
	if !strings.Contains(body, "OURS-EDITED-NOTE") || strings.Contains(body, "ORIGINAL-NOTE") {
		t.Fatalf("our note edit must win over an untouched side:\n%s", body)
	}
}

// Scenario: resolve refuses when both sides edited the Backlog phase note
func TestRsh_ResolveRefusesBothSidesShellEdit(t *testing.T) {
	bin := buildCent(t)
	dir := rshConflictRepo(t,
		shellDoc("ORIGINAL-NOTE", ""),
		shellDoc("OURS-EDITED-NOTE", ""),
		shellDoc("THEIRS-EDITED-NOTE", ""))

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code == 0 {
		t.Fatalf("two different note edits must not be auto-merged:\n%s", out)
	}
	containsAll(t, out, "Backlog", "changed on both sides")
	if !rshUnmerged(t, dir) {
		t.Fatal("the conflict must be left unresolved in the index")
	}
	if !strings.Contains(mustRead(t, dir+"/.workflow/roadmap.json"), "<<<<<<<") {
		t.Fatal("the conflict markers must be left in place")
	}
}
