package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-json-contract.feature
// Scenario: Missing roadmap.json fails roadmap --json with a stderr error and no stdout JSON
// Scenario: Missing roadmap.json fails roadmap ready --json with a stderr error and no stdout JSON
// Scenario: Missing roadmap.json fails roadmap show --json with a stderr error and no stdout JSON
// Scenario: Missing roadmap.json also fails the text-mode commands the same way
func TestRoadmapMissingFileFailsAllSurfaces(t *testing.T) {
	d := rmcProject(t, "") // no roadmap.json written
	surfaces := [][]string{
		{"roadmap", "--json"},
		{"roadmap", "ready", "--json"},
		{"roadmap", "show", "--json"},
		{"roadmap"},
	}
	for _, args := range surfaces {
		so, se, code := rmcRun(t, d, args...)
		if code == 0 {
			t.Fatalf("%v must exit non-zero on missing roadmap", args)
		}
		if strings.TrimSpace(se) == "" {
			t.Fatalf("%v must print an error to stderr", args)
		}
		if strings.ContainsAny(strings.TrimSpace(so), "{[") {
			t.Fatalf("%v must emit no partial stdout JSON, got %q", args, so)
		}
	}
}

// Acceptance: specs/roadmap-json-contract.feature
// Scenario: ready --json when nothing is ready emits an empty array, not null
func TestRoadmapReadyJSONNothingReadyEmptyArray(t *testing.T) {
	// "a" is in-progress and "b" depends on it → neither is ready.
	d := rmcProject(t, `{"phases":[{"name":"Q1","features":[{"name":"a"},{"name":"b","dependsOn":["a"]}]}]}`)
	rmcSeed(t, d, "a", "code")
	out, _, code := rmcRun(t, d, "roadmap", "ready", "--json")
	if code != 0 {
		t.Fatalf("ready --json exit=%d", code)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Fatalf("nothing-ready must emit [] not null, got %q", out)
	}
}

// Acceptance: specs/roadmap-json-contract.feature
// Scenario: Malformed roadmap.json (invalid JSON) is rejected by Load, not partially rendered
// Scenario: Dependency-cycle roadmap.json is rejected by Load, not partially rendered
func TestRoadmapRejectsMalformedAndCyclicSource(t *testing.T) {
	bad := map[string]string{
		"malformed": `{"phases":[`,
		"cycle": `{"phases":[{"name":"Q1","features":[` +
			`{"name":"a","dependsOn":["b"]},{"name":"b","dependsOn":["a"]}]}]}`,
	}
	for label, body := range bad {
		d := rmcProject(t, body)
		so, se, code := rmcRun(t, d, "roadmap", "--json")
		if code == 0 {
			t.Fatalf("%s roadmap.json must be rejected by Load", label)
		}
		if strings.TrimSpace(se) == "" {
			t.Fatalf("%s must print an error to stderr", label)
		}
		if strings.ContainsAny(strings.TrimSpace(so), "{[") {
			t.Fatalf("%s must emit no partial stdout JSON, got %q", label, so)
		}
	}
}
