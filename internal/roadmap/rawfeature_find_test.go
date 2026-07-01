package roadmap

import (
	"strings"
	"testing"
)

// docFrom builds a rawDoc from a JSON body written to a temp file.
func docFrom(t *testing.T, body string) *rawDoc {
	t.Helper()
	doc, err := readRawRoadmap(crudWrite(t, body))
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	return doc
}

// TestFindFeature locates a feature and reports its phase/feature indices.
func TestFindFeature(t *testing.T) {
	doc := docFrom(t, crudBody)
	raw, pi, fi, err := doc.findFeature("billing-api")
	if err != nil || pi != 1 || fi != 0 || len(raw) == 0 {
		t.Fatalf("findFeature billing-api: pi=%d fi=%d err=%v", pi, fi, err)
	}
	if _, _, _, err := doc.findFeature("ghost"); err == nil {
		t.Fatal("missing feature must error")
	}
}

// TestFeaturePhase returns the owning phase name.
func TestFeaturePhase(t *testing.T) {
	doc := docFrom(t, crudBody)
	name, err := doc.featurePhase("checkout-ui")
	if err != nil || name != "Phase 1: Foundations" {
		t.Fatalf("featurePhase: %q err=%v", name, err)
	}
	if _, err := doc.featurePhase("ghost"); err == nil {
		t.Fatal("missing feature must error")
	}
}

// TestToRoadmap decodes mutated phases into a typed Roadmap.
func TestToRoadmap(t *testing.T) {
	doc := docFrom(t, crudBody)
	r, err := doc.toRoadmap()
	if err != nil {
		t.Fatalf("toRoadmap: %v", err)
	}
	if len(r.Phases) != 4 || r.Phases[0].Name != "Phase 1: Foundations" {
		t.Fatalf("unexpected typed roadmap: %+v", r.Phases)
	}
}

// TestFeatureDependents scans all phases, drafts included.
func TestFeatureDependents(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1","features":[{"name":"a"},` +
		`{"name":"b","dependsOn":["a"]},{"name":"c","dependsOn":["a"],"draft":true}]}]}`
	deps, err := docFrom(t, body).featureDependents("a")
	if err != nil || strings.Join(deps, ",") != "b,c" {
		t.Fatalf("dependents of a must be [b c] incl draft: %v err=%v", deps, err)
	}
	none, _ := docFrom(t, body).featureDependents("b")
	if len(none) != 0 {
		t.Fatalf("b has no dependents: %v", none)
	}
}
