// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// mdBase/mdOurs/mdTheirs are the modify-vs-delete conflict: ours edits the
// finding's summary, theirs removes the finding entirely.
const mdBase = `{"phases":[{"name":"Phase 1","features":[{"name":"alpha"}]},
 {"name":"Backlog","features":[
  {"name":"finding-A","summary":"original","deferredAt":"2026-01-01T00:00:00Z"}]}]}`

const mdOurs = `{"phases":[{"name":"Phase 1","features":[{"name":"alpha"}]},
 {"name":"Backlog","features":[
  {"name":"finding-A","summary":"IMPORTANT-NEW-DETAIL","deferredAt":"2026-01-01T00:00:00Z"}]}]}`

const mdTheirs = `{"phases":[{"name":"Phase 1","features":[{"name":"alpha"}]},
 {"name":"Backlog","features":[]}]}`

// Scenario: resolve refuses a finding one side edited and the other deleted
//
// Deciding by presence alone answered this with "the deletion wins", which
// threw the edit away with exit 0 and a `kept 0 findings` summary — the precise
// silent data loss `roadmap resolve` exists to prevent, on a smaller object
// than a phase.
func TestRsh_ResolveRefusesModifyDelete(t *testing.T) {
	bin := buildCent(t)
	dir := rshConflictRepo(t, mdBase, mdOurs, mdTheirs)
	before := mustRead(t, dir+"/.workflow/roadmap.json")
	stagedBefore := rshGitOut(t, dir, "ls-files", "-s", "--", ".workflow/roadmap.json")

	out, code := runCent(t, bin, dir, "roadmap", "resolve")
	if code == 0 {
		t.Fatalf("a modify/delete pair must NOT resolve to exit 0:\n%s", out)
	}
	containsAll(t, out, "finding-A", "modified on one side and deleted")

	after := mustRead(t, dir+"/.workflow/roadmap.json")
	if after != before {
		t.Fatalf("a refusal must leave roadmap.json byte-identical:\n%s", after)
	}
	if !strings.Contains(after, "<<<<<<<") {
		t.Fatal("the conflict markers must be left in place")
	}
	if !strings.Contains(after, "IMPORTANT-NEW-DETAIL") {
		t.Fatal("the edited content must still be recoverable from the conflict")
	}
	if !rshUnmerged(t, dir) {
		t.Fatal("roadmap.json must still be unmerged in the index")
	}
	if got := rshGitOut(t, dir, "ls-files", "-s", "--", ".workflow/roadmap.json"); got != stagedBefore {
		t.Fatalf("the index must be byte-identical across the refusal:\n%s", got)
	}
}
