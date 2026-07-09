package integration_test

// Acceptance: specs/roadmap-phase-ops.feature
// Scenario: phase remove --force removes the phase, its features, and their analysis/quality entries, then validate PASSes
// Scenario: phase remove --force is REFUSED byte-identical when a surviving feature depends on a removed one
// Scenario: removing a middle phase while mutating a later feature reindexes the dirty map so both renders are correct

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// poProject chdirs into a temp project seeded with body plus analysis/quality
// artifacts (entries for auth-service/billing-api/reporting) and markdown companions.
func poProject(t *testing.T, body string) {
	t.Helper()
	intoProject(t, body)
	entry := func(n string) string {
		return `{"name":"` + n + `","scores":{"acceptanceCriteria":9,"userValue":9,` +
			`"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}`
	}
	os.WriteFile(roadmap.RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[`+ //nolint:errcheck
		`{"name":"auth-service","summary":"s"},{"name":"billing-api","summary":"s"},{"name":"reporting","summary":"s"}]}`), 0o644)
	os.WriteFile(roadmap.RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[`+ //nolint:errcheck
		entry("auth-service")+`,`+entry("billing-api")+`,`+entry("reporting")+`]}`), 0o644)
}

const poIntBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"},{"name":"reporting"}]}]}`

// TestPO_ForcePrunesAndValidates crosses force-remove → prune → validate PASS.
func TestPO_ForcePrunesAndValidates(t *testing.T) {
	poProject(t, poIntBody)
	if err := roadmap.PhaseRemove(roadmap.RoadmapFile, "Phase 2: Growth", true); err != nil {
		t.Fatalf("force remove: %v", err)
	}
	for _, f := range []string{roadmap.RoadmapFile, roadmap.RoadmapAnalysisFile, roadmap.RoadmapQualityFile} {
		b, _ := os.ReadFile(f)
		if strings.Contains(string(b), "billing-api") || strings.Contains(string(b), "reporting") {
			t.Fatalf("%s must be pruned", f)
		}
	}
	r, _ := roadmap.Load()
	if err := roadmap.ValidateAnalysis(r); err != nil {
		t.Fatalf("ValidateAnalysis must PASS: %v", err)
	}
	if err := roadmap.ValidateQuality(r); err != nil {
		t.Fatalf("ValidateQuality must PASS: %v", err)
	}
}

// TestPO_ForceRefusedOnSurvivingDep refuses byte-identical across all three files.
func TestPO_ForceRefusedOnSurvivingDep(t *testing.T) {
	body := `{"phases":[` +
		`{"name":"Phase 1: Foundations","features":[{"name":"auth-service","dependsOn":["billing-api"]}]},` +
		`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]}]}`
	poProject(t, body)
	before := map[string][]byte{}
	files := []string{roadmap.RoadmapFile, roadmap.RoadmapAnalysisFile, roadmap.RoadmapQualityFile}
	for _, f := range files {
		before[f], _ = os.ReadFile(f)
	}
	if err := roadmap.PhaseRemove(roadmap.RoadmapFile, "Phase 2: Growth", true); err == nil ||
		!strings.Contains(err.Error(), "depends on") {
		t.Fatalf("must refuse with depends-on error: %v", err)
	}
	for _, f := range files {
		after, _ := os.ReadFile(f)
		if !bytes.Equal(before[f], after) {
			t.Fatalf("%s must be byte-identical", f)
		}
	}
}
