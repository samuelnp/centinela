package roadmap

import (
	"os"
	"strings"
	"testing"
)

// TestAppendPromotionArtifacts_WritesAllFour writes all four artifact files.
func TestAppendPromotionArtifacts_WritesAllFour(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })                                                                               //nolint:errcheck
	os.Chdir(d)                                                                                                        //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                     //nolint:errcheck
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[]}`), 0644)                 //nolint:errcheck
	os.WriteFile(RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0644) //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis\n"), 0644)                                                //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality\n"), 0644)                                                  //nolint:errcheck
	scores, _ := ParseScores("9,9,9,9,9,9")
	f := &BacklogFinding{
		Name:       "new-slug",
		Summary:    "a finding",
		Source:     &Source{Feature: "dfrc", Role: "eng"},
		DeferredAt: "2026-01-01T00:00:00Z",
	}
	if err := appendPromotionArtifacts("new-slug", "a finding", scores, f); err != nil {
		t.Fatalf("appendPromotionArtifacts: %v", err)
	}
	analysisData, _ := os.ReadFile(RoadmapAnalysisFile)
	if !strings.Contains(string(analysisData), "new-slug") {
		t.Error("analysis must contain promoted slug")
	}
	qualityData, _ := os.ReadFile(RoadmapQualityFile)
	if !strings.Contains(string(qualityData), "new-slug") {
		t.Error("quality must contain promoted slug")
	}
	analysisMd, _ := os.ReadFile(RoadmapAnalysisMarkdown)
	if !strings.Contains(string(analysisMd), "new-slug") {
		t.Error("analysis.md must contain provenance bullet")
	}
	qualityMd, _ := os.ReadFile(RoadmapQualityMarkdown)
	if !strings.Contains(string(qualityMd), "new-slug") {
		t.Error("quality.md must contain provenance bullet")
	}
}

// TestAppendFeatureEntry_EmptyFeaturesKey handles missing features key.
func TestAppendFeatureEntry_EmptyFeaturesKey(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(d)                          //nolint:errcheck
	os.MkdirAll(".workflow", 0755)       //nolint:errcheck
	// No "features" key — must create it
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"pm"}`), 0644) //nolint:errcheck
	entry, _ := compactBytes(AnalysisFeature{Name: "slug"})
	if err := appendFeatureEntry(RoadmapAnalysisFile, entry); err != nil {
		t.Fatalf("appendFeatureEntry with no features key: %v", err)
	}
	data, _ := os.ReadFile(RoadmapAnalysisFile)
	if !strings.Contains(string(data), "slug") {
		t.Errorf("slug must appear when features key was absent: %s", data)
	}
}
