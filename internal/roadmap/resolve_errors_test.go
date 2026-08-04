package roadmap

import (
	"encoding/json"
	"strings"
	"testing"
)

// Every unmarshalable region must surface as an error rather than a partial
// merge. Each case here is a distinct decode site inside Resolve.
func TestResolveRefusesEveryUndecodableRegion(t *testing.T) {
	good := backlogDoc(finding("a", "2026-01-01T00:00:00Z"))
	cases := map[string]string{
		"backlog features not an array":   `{"phases":[{"name":"Backlog","features":3}]}`,
		"backlog feature not an object":   `{"phases":[{"name":"Backlog","features":["oops"]}]}`,
		"schedulable phase not an object": `{"phases":[42]}`,
		"top level not an object":         `[]`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Resolve(good, []byte(body), good); err == nil {
				t.Fatal("want a refusal")
			}
			if _, err := Resolve(good, good, []byte(body)); err == nil {
				t.Fatal("want a refusal on their side too")
			}
		})
	}
}

// A base whose Backlog is undecodable is refused before any merge decision.
func TestResolveRefusesAnUndecodableBase(t *testing.T) {
	good := backlogDoc(finding("a", "2026-01-01T00:00:00Z"))
	bad := []byte(`{"phases":[{"name":"Backlog","features":{"x":1}}]}`)
	if _, err := Resolve(bad, good, good); err == nil {
		t.Fatal("an undecodable base must be refused")
	}
}

// compactJSON is the whitespace normalizer the three-way comparison rests on.
func TestCompactJSONNormalizesAndRefusesGarbage(t *testing.T) {
	got, err := compactJSON(json.RawMessage("{\n  \"a\" : 1\n}"))
	if err != nil || string(got) != `{"a":1}` {
		t.Fatalf("compactJSON = %q, %v", got, err)
	}
	if _, err := compactJSON(json.RawMessage("{oops")); err == nil {
		t.Fatal("invalid JSON must be refused")
	}
}

// The nameless-entry refusal names the side, so an operator knows where to look.
func TestResolveNamesTheSideOfANamelessFinding(t *testing.T) {
	good := backlogDoc()
	nameless := []byte(`{"phases":[{"name":"Backlog","features":[{"summary":"x"}]}]}`)
	_, err := Resolve(good, good, nameless)
	if err == nil || !strings.Contains(err.Error(), "their side") {
		t.Fatalf("want a named refusal, got %v", err)
	}
}
