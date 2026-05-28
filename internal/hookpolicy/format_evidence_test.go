package hookpolicy

import (
	"bytes"
	"strings"
	"testing"
)

const minified = `{"feature":"alpha","step":"plan","role":"big-thinker","status":"done",` +
	`"generatedAt":"2026-05-12T00:00:00Z","inputs":[],"outputs":[],"edgeCases":[],"handoffTo":"feature-specialist"}`

func TestFormatEvidencePrettyPrintsActive(t *testing.T) {
	out, changed, err := FormatEvidence(".workflow/alpha-big-thinker.json", []byte(minified), "alpha")
	if err != nil || !changed {
		t.Fatalf("expected changed=true, got changed=%v err=%v", changed, err)
	}
	if !bytes.Contains(out, []byte("\n  \"feature\": \"alpha\"")) {
		t.Fatalf("not pretty-printed:\n%s", out)
	}
}

func TestFormatEvidenceLeavesOtherFeatures(t *testing.T) {
	out, changed, _ := FormatEvidence(".workflow/beta-big-thinker.json", []byte(minified), "alpha")
	if changed || !bytes.Equal(out, []byte(minified)) {
		t.Fatal("other-feature file should be untouched")
	}
}

func TestFormatEvidenceNonJSONPassthrough(t *testing.T) {
	body := []byte("# not json\n")
	out, changed, _ := FormatEvidence(".workflow/alpha-edge-cases.md", body, "alpha")
	if changed || !bytes.Equal(out, body) {
		t.Fatal(".md should be untouched")
	}
}

func TestFormatEvidenceParseFailurePassthrough(t *testing.T) {
	body := []byte(`{not really json`)
	out, changed, _ := FormatEvidence(".workflow/alpha-big-thinker.json", body, "alpha")
	if changed || !bytes.Equal(out, body) {
		t.Fatal("parse failure should be a silent passthrough")
	}
}

func TestFormatEvidenceIdempotent(t *testing.T) {
	first, _, _ := FormatEvidence(".workflow/alpha-big-thinker.json", []byte(minified), "alpha")
	second, changed, _ := FormatEvidence(".workflow/alpha-big-thinker.json", first, "alpha")
	if changed || !bytes.Equal(first, second) {
		t.Fatal("formatter is not idempotent")
	}
}

func TestFormatEvidenceEmptyFeatureSkipsAll(t *testing.T) {
	out, changed, _ := FormatEvidence(".workflow/alpha-big-thinker.json", []byte(minified), "")
	if changed || !bytes.Equal(out, []byte(minified)) {
		t.Fatal("no active feature should be a no-op")
	}
}

func TestFormatEvidencePreservesUnknownFields(t *testing.T) {
	withExtra := `{"feature":"alpha","step":"plan","role":"big-thinker","status":"done",` +
		`"generatedAt":"x","inputs":[],"outputs":[],"edgeCases":[],"handoffTo":"x","extra_legacy":"keep"}`
	out, changed, _ := FormatEvidence(".workflow/alpha-big-thinker.json", []byte(withExtra), "alpha")
	if !changed || !strings.Contains(string(out), `"extra_legacy": "keep"`) {
		t.Fatalf("unknown field dropped:\n%s", out)
	}
}
