package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

func setupArtifacts(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })                                                                               //nolint:errcheck
	os.Chdir(d)                                                                                                        //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                     //nolint:errcheck
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[]}`), 0644)                 //nolint:errcheck
	os.WriteFile(RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0644) //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis"), 0644)                                                  //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality"), 0644)                                                    //nolint:errcheck
	return d
}

// TestPreflightArtifacts_AllPresent passes when all four files exist.
func TestPreflightArtifacts_AllPresent(t *testing.T) {
	setupArtifacts(t)
	if err := preflightArtifacts(); err != nil {
		t.Fatalf("preflightArtifacts: %v", err)
	}
}

// TestPreflightArtifacts_MissingAnalysisJSON fails correctly.
func TestPreflightArtifacts_MissingAnalysisJSON(t *testing.T) {
	setupArtifacts(t)
	os.Remove(RoadmapAnalysisFile) //nolint:errcheck
	err := preflightArtifacts()
	if err == nil {
		t.Fatal("expected error for missing analysis JSON")
	}
}

// TestPreflightArtifacts_MissingQualityJSON fails correctly.
func TestPreflightArtifacts_MissingQualityJSON(t *testing.T) {
	setupArtifacts(t)
	os.Remove(RoadmapQualityFile) //nolint:errcheck
	if err := preflightArtifacts(); err == nil {
		t.Fatal("expected error for missing quality JSON")
	}
}

// TestPreflightArtifacts_MissingAnalysisMd fails correctly.
func TestPreflightArtifacts_MissingAnalysisMd(t *testing.T) {
	setupArtifacts(t)
	os.Remove(RoadmapAnalysisMarkdown) //nolint:errcheck
	if err := preflightArtifacts(); err == nil {
		t.Fatal("expected error for missing analysis markdown")
	}
}

// TestPreflightArtifacts_MissingQualityMd fails correctly.
func TestPreflightArtifacts_MissingQualityMd(t *testing.T) {
	setupArtifacts(t)
	os.Remove(RoadmapQualityMarkdown) //nolint:errcheck
	if err := preflightArtifacts(); err == nil {
		t.Fatal("expected error for missing quality markdown")
	}
}

// TestCheckArtifactJSON_CorruptJSON returns error for invalid JSON.
func TestCheckArtifactJSON_CorruptJSON(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bad.json")
	os.WriteFile(p, []byte("{bad"), 0644) //nolint:errcheck
	if err := checkArtifactJSON(p); err == nil {
		t.Error("expected error for corrupt JSON")
	}
}

// TestCheckArtifactJSON_InvalidFeaturesArray returns error when features is not array.
func TestCheckArtifactJSON_InvalidFeaturesArray(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bad.json")
	os.WriteFile(p, []byte(`{"features":"not-array"}`), 0644) //nolint:errcheck
	if err := checkArtifactJSON(p); err == nil {
		t.Error("expected error for non-array features")
	}
}
