package roadmap

import (
	"strings"
	"testing"
)

// Unknown top-level, per-phase and per-feature fields must survive verbatim.
func TestCanonicalRenderPreservesUnknownFields(t *testing.T) {
	body := `{"intro":"hi","future":{"k":1},"phases":[` +
		`{"name":"P","owner":"team-a","features":[{"name":"a","futureField":["x"]}]}]}`
	out := renderStr(t, docFrom(t, body))
	for _, want := range []string{`"k": 1`, `"owner": "team-a"`, `{"name":"a","futureField":["x"]}`, `"intro": "hi"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s lost in the round trip:\n%s", want, out)
		}
	}
}

// A phase with no "features" key must still render valid, canonical JSON.
func TestCanonicalRenderPhaseWithoutFeaturesKey(t *testing.T) {
	out := renderStr(t, docFrom(t, `{"phases":[{"name":"Empty"},{"name":"P","features":[{"name":"a"}]}]}`))
	if !strings.Contains(out, `"features": []`) {
		t.Fatalf("a phase with no features key must normalize to an empty array:\n%s", out)
	}
	if _, err := parseRawRoadmap([]byte(out)); err != nil {
		t.Fatalf("render produced invalid JSON: %v\n%s", err, out)
	}
}
