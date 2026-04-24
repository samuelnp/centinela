package setup

import (
	"encoding/json"
	"testing"
)

func TestMergeInstructionsPreservesCustomOrder(t *testing.T) {
	raw := map[string]json.RawMessage{}
	raw["instructions"], _ = json.Marshal([]string{"RULES.md", "TEAM.md", "RULES.md", "AGENTS.md"})

	if !mergeInstructions(raw) {
		t.Fatal("expected mergeInstructions to report a change")
	}

	var got []string
	json.Unmarshal(raw["instructions"], &got) //nolint:errcheck
	want := []string{"RULES.md", "TEAM.md", "AGENTS.md", "CLAUDE.md"}
	if !sameStrings(got, want) {
		t.Fatalf("unexpected instructions: %#v", got)
	}
}

func TestSameStringsAndHasValue(t *testing.T) {
	if !sameStrings([]string{"a", "b"}, []string{"a", "b"}) {
		t.Fatal("expected identical slices to match")
	}
	if sameStrings([]string{"a"}, []string{"b"}) {
		t.Fatal("expected different slices not to match")
	}
	if !hasValue([]string{"a", "b"}, "b") {
		t.Fatal("expected value lookup to succeed")
	}
	if hasValue([]string{"a", "b"}, "c") {
		t.Fatal("expected missing value lookup to fail")
	}
}
