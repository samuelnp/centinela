package hookpolicy

import (
	"bytes"
	"strings"
	"testing"
)

func TestIsActiveFeatureEvidenceFiltersDirectories(t *testing.T) {
	// Non-`.workflow/` paths are out of scope even if the basename matches.
	out, changed, _ := FormatEvidence("/tmp/other/alpha-big-thinker.json", []byte(minified), "alpha")
	if changed || !bytes.Equal(out, []byte(minified)) {
		t.Fatal("path outside .workflow should be skipped")
	}
}

func TestFormatEvidenceRejectsNonObjectJSON(t *testing.T) {
	// A JSON array is parseable but does not match `map[string]json.RawMessage`.
	body := []byte(`[1,2,3]`)
	out, changed, _ := FormatEvidence(".workflow/alpha-big-thinker.json", body, "alpha")
	if changed || !bytes.Equal(out, body) {
		t.Fatal("non-object JSON should fall through unchanged")
	}
}

func TestEncodeOrderedEvidencePreservesUnknownSorted(t *testing.T) {
	// Constructed object with two extras out of alphabetical order should
	// come back sorted.
	body := `{"zeta":1,"alpha":2,"feature":"x","step":"y","role":"z","status":"a",` +
		`"generatedAt":"a","inputs":[],"outputs":[],"edgeCases":[],"handoffTo":"a"}`
	out, _, _ := FormatEvidence(".workflow/x-big-thinker.json", []byte(body), "x")
	a := strings.Index(string(out), `"alpha"`)
	z := strings.Index(string(out), `"zeta"`)
	if a == -1 || z == -1 || a > z {
		t.Fatalf("extras not sorted: %s", out)
	}
}
