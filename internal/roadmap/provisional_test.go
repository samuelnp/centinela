package roadmap

import (
	"os"
	"strings"
	"testing"
)

// TestSeededArtifactsAreRefusedAsProvisional is the cross-profile guard: an
// empty roadmap is otherwise fully COVERED by an empty artifact pair, so
// without the provisional mark one guided promote would permanently satisfy the
// strict grading rung with no evaluator pass. Both validators must refuse, and
// both must say why and how to clear it.
func TestSeededArtifactsAreRefusedAsProvisional(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("SeedArtifactsIfAbsent: %v", err)
	}
	empty := &Roadmap{}
	for name, err := range map[string]error{
		"analysis": ValidateAnalysis(empty),
		"quality":  ValidateQuality(empty),
	} {
		if err == nil {
			t.Fatalf("seeded %s must not satisfy a grading rung", name)
		}
		if !strings.Contains(err.Error(), "provisional") ||
			!strings.Contains(err.Error(), "no ") {
			t.Fatalf("seeded %s refusal must name the mark and the missing pass: %v", name, err)
		}
	}
}

// TestClearingProvisionalRestoresTheArtifact: the mark is the ONLY thing
// refusing the seeded pair, so removing it (what a real evaluator pass does)
// must make the same files valid. Otherwise the refusal would be a dead end.
func TestClearingProvisionalRestoresTheArtifact(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("SeedArtifactsIfAbsent: %v", err)
	}
	for _, p := range []string{RoadmapAnalysisFile, RoadmapQualityFile} {
		body, _ := os.ReadFile(p)
		cleaned := strings.Replace(string(body), `  "provisional": true,`+"\n", "", 1)
		os.WriteFile(p, []byte(cleaned), 0644) //nolint:errcheck
	}
	empty := &Roadmap{}
	if err := ValidateAnalysis(empty); err != nil {
		t.Fatalf("analysis must validate once the mark is cleared: %v", err)
	}
	if err := ValidateQuality(empty); err != nil {
		t.Fatalf("quality must validate once the mark is cleared: %v", err)
	}
}

// TestSeedThenPreflightRespectsTheCallerSwitch: the seam both promote branches
// share. With seeding off (strict) an absent artifact still refuses; with it on
// (guided) the same tree preflights clean.
func TestSeedThenPreflightRespectsTheCallerSwitch(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := seedThenPreflight(PromoteRequest{}); err == nil {
		t.Fatal("without seeding, a missing artifact must still refuse")
	}
	for _, p := range seededPaths {
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("a non-seeding request must create nothing, found %s", p)
		}
	}
	if err := seedThenPreflight(PromoteRequest{SeedArtifacts: true}); err != nil {
		t.Fatalf("with seeding, the same tree must preflight clean: %v", err)
	}
	if err := seedThenPreflight(PromoteRequest{SeedArtifacts: true}); err != nil {
		t.Fatalf("seeding again must stay clean: %v", err)
	}
	// A seed that cannot be written surfaces instead of falling through to a
	// preflight that would then blame a "missing" artifact.
	t.Chdir(t.TempDir())
	os.WriteFile(".workflow", []byte("not a dir"), 0644) //nolint:errcheck
	if err := seedThenPreflight(PromoteRequest{SeedArtifacts: true}); err == nil {
		t.Fatal("an unwritable seed path must surface")
	}
}
