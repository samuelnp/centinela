package roadmap

import (
	"encoding/json"
	"strings"
	"testing"
)

// sideWith builds a side holding one raw phase body, bypassing parseSide so a
// shape parseSide would already have rejected can still be exercised here.
func sideWith(label, name, body string) *side {
	return &side{
		label: label,
		order: []string{name},
		phase: map[string]json.RawMessage{name: json.RawMessage(body)},
		rest:  map[string]json.RawMessage{},
	}
}

// A phase body that is not an object must be refused BY SIDE, never merged as
// an empty shell — silently returning {} would drop every key it carried.
func TestPhaseObjectRefusesANonObjectPhase(t *testing.T) {
	for _, label := range []string{"the base", "our side", "their side"} {
		s := sideWith(label, "Backlog", `["not","an","object"]`)
		_, err := s.phaseObject("Backlog")
		if err == nil {
			t.Fatalf("%s: a non-object phase must be refused", label)
		}
		if !strings.Contains(err.Error(), label) || !strings.Contains(err.Error(), "Backlog") {
			t.Fatalf("the error must name the side and the phase: %v", err)
		}
	}
}

// A side that simply does not have the phase yields an empty shell, not an error.
func TestPhaseObjectIsEmptyWhenTheSideLacksThePhase(t *testing.T) {
	got, err := sideWith("our side", "Backlog", `{"name":"Backlog"}`).phaseObject("Other")
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

// Each of the three sides is decoded, so a malformed one on ANY side stops the
// merge rather than producing a shell built from the other two.
func TestMergeShellKeysRefusesAMalformedSide(t *testing.T) {
	good := sideWith("good", "Backlog", `{"name":"Backlog","note":"n"}`)
	bad := sideWith("bad side", "Backlog", `"just a string"`)
	cases := map[string][3]*side{
		"base":   {bad, good, good},
		"ours":   {good, bad, good},
		"theirs": {good, good, bad},
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := mergeShellKeys("Backlog", s[0], s[1], s[2]); err == nil {
				t.Fatal("a malformed side must stop the shell merge")
			}
		})
	}
}

// The shell merge never carries `features` over from a side: it is always the
// merged finding list the caller writes afterwards.
func TestMergeShellKeysNeverCarriesFeatures(t *testing.T) {
	b := sideWith("the base", "Backlog", `{"name":"Backlog","features":[{"name":"stale"}]}`)
	o := sideWith("our side", "Backlog", `{"name":"Backlog","features":[{"name":"stale"}]}`)
	tt := sideWith("their side", "Backlog", `{"name":"Backlog","features":[{"name":"other"}]}`)
	got, err := mergeShellKeys("Backlog", b, o, tt)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got[shellKey]; ok {
		t.Fatalf("features must not survive the shell merge: %v", got)
	}
	if string(got["name"]) != `"Backlog"` {
		t.Fatalf("the phase name must survive: %v", got)
	}
}
