package orchestration

import "testing"

// The bypass StubEntry exists to close: blanking the scaffold's descriptions to
// ellipses turns every slot into a "citation", which UnreplacedSlot passes.
func TestStubEntry_RejectsAllCitationScaffold(t *testing.T) {
	stubs := []string{
		"- <FILL: ...>",
		"- <FILL: ...>: <FILL: ...>",
		"<FILL: ...>",
		"- <FILL: >",
		"* <FILL: …>",
		"- <FILL: type>: <FILL: one-line summary of the change>",
		"- fix: <FILL: one-line summary>",
	}
	for _, s := range stubs {
		if !StubEntry(s) {
			t.Errorf("%q is scaffold, not a written entry", s)
		}
	}
}

// Prose that says something either side of the marker is a quotation, and must
// keep passing — including the line this feature itself ships.
func TestStubEntry_AcceptsProseQuotingTheMarker(t *testing.T) {
	written := []string{
		"- fix: reject an unreplaced <FILL: ...> marker behind a preamble",
		"- feat: bind the evidence gates so a stub can no longer pass",
		"the <FILL: ...> form is how the docs cite the marker",
		"- docs: explain <FILL: ...> and why it is refused",
	}
	for _, s := range written {
		if StubEntry(s) {
			t.Errorf("%q is written prose and must pass", s)
		}
	}
}

// StubEntry must not widen UnreplacedSlot for entries carrying no slot at all.
func TestStubEntry_IgnoresEntriesWithoutSlots(t *testing.T) {
	for _, s := range []string{"- fix: something real", "## Changelog", "-", ""} {
		if StubEntry(s) {
			t.Errorf("%q carries no slot and must not read as scaffold", s)
		}
	}
}
