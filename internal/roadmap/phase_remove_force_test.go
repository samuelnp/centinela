package roadmap

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// forceBody: Phase 2 holds two scored features that also have analysis/quality
// entries; Phase 1 and Backlog survive.
const forceBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"},{"name":"reporting"}]},` +
	`{"name":"Backlog","features":[]}]}`

// seedArtifacts writes analysis+quality entries for billing-api and reporting.
func seedArtifacts(t *testing.T) {
	t.Helper()
	entry := func(n string) string {
		return `{"name":"` + n + `","scores":{"acceptanceCriteria":9,"userValue":9,` +
			`"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}`
	}
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[`+ //nolint:errcheck
		`{"name":"auth-service","summary":"s"},{"name":"billing-api","summary":"s"},{"name":"reporting","summary":"s"}]}`), 0o644)
	os.WriteFile(RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[`+ //nolint:errcheck
		entry("auth-service")+`,`+entry("billing-api")+`,`+entry("reporting")+`]}`), 0o644)
}

// TestPhaseRemove_ForcePrunesAndValidates: --force drops the phase, its features,
// and their analysis+quality entries, then analysis/quality validation PASSes.
func TestPhaseRemove_ForcePrunesAndValidates(t *testing.T) {
	phaseOpsChdir(t, forceBody)
	seedArtifacts(t)
	if err := PhaseRemove(RoadmapFile, "Phase 2: Growth", true); err != nil {
		t.Fatalf("force remove: %v", err)
	}
	road := string(crudBytes(t, RoadmapFile))
	if strings.Contains(road, "billing-api") || strings.Contains(road, "reporting") || strings.Contains(road, "Phase 2: Growth") {
		t.Fatalf("phase+features must be gone: %s", road)
	}
	for _, f := range []string{RoadmapAnalysisFile, RoadmapQualityFile} {
		b := string(crudBytes(t, f))
		if strings.Contains(b, "billing-api") || strings.Contains(b, "reporting") {
			t.Fatalf("%s must be pruned: %s", f, b)
		}
	}
	r, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := ValidateAnalysis(r); err != nil {
		t.Fatalf("ValidateAnalysis must PASS: %v", err)
	}
	if err := ValidateQuality(r); err != nil {
		t.Fatalf("ValidateQuality must PASS: %v", err)
	}
}

// TestPhaseRemove_ForceRefusedOnSurvivingDep: a surviving feature dependsOn a removed
// one → refused, all three files byte-identical.
func TestPhaseRemove_ForceRefusedOnSurvivingDep(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1: Foundations","features":[{"name":"checkout-ui","dependsOn":["billing-api"]}]},` +
		`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]}]}`
	phaseOpsChdir(t, body)
	seedArtifacts(t)
	before := map[string][]byte{}
	for _, f := range []string{RoadmapFile, RoadmapAnalysisFile, RoadmapQualityFile} {
		before[f] = crudBytes(t, f)
	}
	wantErr(t, PhaseRemove(RoadmapFile, "Phase 2: Growth", true), "depends on")
	for _, f := range []string{RoadmapFile, RoadmapAnalysisFile, RoadmapQualityFile} {
		if !bytes.Equal(before[f], crudBytes(t, f)) {
			t.Fatalf("%s must be byte-identical after refused force remove", f)
		}
	}
}
