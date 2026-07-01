package roadmap

import (
	"strings"
	"testing"
)

// TestRewriteDependents_MultiPhase rewrites every dependent's dependsOn old→new
// across phases, preserving non-matching entries and untouched phases.
func TestRewriteDependents_MultiPhase(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1","features":[{"name":"a"},{"name":"b","dependsOn":["a"]},` +
		`{"name":"c","dependsOn":["a","x"]}]},` +
		`{"name":"Phase 2","features":[{"name":"d","dependsOn":["a"]},{"name":"e"}]},` +
		`{"name":"Phase 3","features":[{"name":"f","dependsOn":["x"]}]}]}`
	doc := docFrom(t, body)
	if err := doc.rewriteDependents("a", "z"); err != nil {
		t.Fatalf("rewriteDependents: %v", err)
	}
	out := renderStr(t, doc)
	for _, want := range []string{
		`{"name":"b","dependsOn":["z"]}`,
		`{"name":"c","dependsOn":["z","x"]}`,
		`{"name":"d","dependsOn":["z"]}`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %s in:\n%s", want, out)
		}
	}
	// Phase 3 has no dependent on "a": it must be untouched (still json.Indent form).
	if !strings.Contains(out, `"f"`) || strings.Contains(out, `{"name":"f"`) {
		t.Fatalf("untouched Phase 3 must not be re-rendered one-per-line:\n%s", out)
	}
	if strings.Contains(out, `"a"`) == false {
		t.Fatal("the renamed target's own entry is left to applyRename, not rewriteDependents")
	}
}

// TestRewriteDependents_NoMatch leaves the doc render-stable when nothing depends
// on the old name.
func TestRewriteDependents_NoMatch(t *testing.T) {
	doc := docFrom(t, crudBody)
	if err := doc.rewriteDependents("nonexistent", "z"); err != nil {
		t.Fatalf("rewriteDependents: %v", err)
	}
	if out := renderStr(t, doc); strings.Contains(out, `"z"`) {
		t.Fatalf("no dependent should be rewritten:\n%s", out)
	}
}
