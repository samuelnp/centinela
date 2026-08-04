// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// ieBase/ieOurs/ieTheirs: THEIRS edits a Backlog finding's summary, OURS leaves
// it alone and appends its own finding (so the merge really conflicts).
const ieBase = `{"phases":[{"name":"Phase 1","features":[{"name":"alpha"}]},
 {"name":"Backlog","features":[
  {"name":"finding-A","summary":"orig","deferredAt":"2026-01-01T00:00:00Z"}]}]}`

const ieOurs = `{"phases":[{"name":"Phase 1","features":[{"name":"alpha"}]},
 {"name":"Backlog","features":[
  {"name":"finding-A","summary":"orig","deferredAt":"2026-01-01T00:00:00Z"},
  {"name":"ours-new","summary":"added by ours","deferredAt":"2026-02-01T00:00:00Z"}]}]}`

const ieTheirs = `{"phases":[{"name":"Phase 1","features":[{"name":"alpha"}]},
 {"name":"Backlog","features":[
  {"name":"finding-A","summary":"EDITED-BY-THEIRS","deferredAt":"2026-01-01T00:00:00Z"},
  {"name":"theirs-new","summary":"added by theirs","deferredAt":"2026-03-01T00:00:00Z"}]}]}`

// Scenario: resolve keeps a one-sided edit made on the incoming side
//
// An edit does not move deferredAt, so ordering the two copies by deferredAt
// tied and silently returned OURS — the base version — while printing a green
// "✓ Resolved roadmap state — kept 3 findings". More reachable than
// modify/delete: ANY edit to an existing Backlog entry on the incoming branch
// was affected.
func TestRsh_ResolveKeepsIncomingOneSidedEdit(t *testing.T) {
	bin := buildCent(t)
	dir := rshConflictRepo(t, ieBase, ieOurs, ieTheirs)

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code != 0 {
		t.Fatalf("a one-sided incoming edit is not a conflict: exit %d\n%s", code, out)
	}
	body := mustRead(t, dir+"/.workflow/roadmap.json")
	if !strings.Contains(body, "EDITED-BY-THEIRS") {
		t.Fatalf("the incoming edit was silently discarded:\n%s\n%s", out, body)
	}
	if strings.Contains(body, `"summary":"orig"`) {
		t.Fatalf("the base version must not survive the incoming edit:\n%s", body)
	}
	// Both sides' additions must still be unioned, unchanged by the fix.
	for _, slug := range []string{"ours-new", "theirs-new"} {
		if !strings.Contains(body, slug) {
			t.Fatalf("%q was dropped:\n%s", slug, body)
		}
	}
	if strings.Contains(body, "<<<<<<<") {
		t.Fatalf("no conflict markers may survive a successful resolve:\n%s", body)
	}
	if rshUnmerged(t, dir) {
		t.Fatal("roadmap.json must be resolved in the index")
	}
}

// Scenario: resolve refuses when both sides edited the same finding differently
//
// Two real edits to one finding cannot be reconciled by rule; refusing is the
// same contract as a both-sides phase edit.
func TestRsh_ResolveRefusesBothSidesFindingEdit(t *testing.T) {
	bin := buildCent(t)
	ours := strings.Replace(ieBase, `"summary":"orig"`, `"summary":"EDITED-BY-OURS"`, 1)
	theirs := strings.Replace(ieBase, `"summary":"orig"`, `"summary":"EDITED-BY-THEIRS"`, 1)
	dir := rshConflictRepo(t, ieBase, ours, theirs)

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code == 0 {
		t.Fatalf("two different edits must not be auto-merged:\n%s", out)
	}
	containsAll(t, out, "finding-A", "modified differently on both sides")
	if !rshUnmerged(t, dir) {
		t.Fatal("the conflict must be left unresolved in the index")
	}
	if !strings.Contains(mustRead(t, dir+"/.workflow/roadmap.json"), "<<<<<<<") {
		t.Fatal("the conflict markers must be left in place")
	}
}
