package roadmap

import (
	"os"
	"strings"
	"testing"
)

func TestValidateAnalysisPassAndMissingFeature(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}, {Name: "post"}}}}}
	os.MkdirAll(".workflow", 0755)                                    //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis"), 0644) //nolint:errcheck
	json := `{"role":"senior-product-manager","features":[{"name":"user","dependsOn":[]},{"name":"post","dependsOn":["user"]}]}`
	os.WriteFile(RoadmapAnalysisFile, []byte(json), 0644) //nolint:errcheck
	if err := ValidateAnalysis(r); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[{"name":"user","dependsOn":[]}]}`), 0644) //nolint:errcheck
	if err := ValidateAnalysis(r); err == nil || !strings.Contains(err.Error(), "analysis missing feature") {
		t.Fatalf("expected missing feature error, got %v", err)
	}
}

func TestValidateAnalysisInvalidBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}}}}}
	os.MkdirAll(".workflow", 0755)                                    //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis"), 0644) //nolint:errcheck
	os.WriteFile(RoadmapAnalysisFile, []byte(`{bad`), 0644)           //nolint:errcheck
	if err := ValidateAnalysis(r); err == nil || !strings.Contains(err.Error(), "invalid roadmap analysis") {
		t.Fatalf("expected invalid json error, got %v", err)
	}
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"qa","features":[{"name":"user"}]}`), 0644) //nolint:errcheck
	if err := ValidateAnalysis(r); err == nil || !strings.Contains(err.Error(), "senior-product-manager") {
		t.Fatalf("expected role error, got %v", err)
	}
	// Option B: analysis no longer carries dependsOn; it validates feature
	// coverage only. A feature not present on the roadmap is still rejected.
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[{"name":"ghost"}]}`), 0644) //nolint:errcheck
	if err := ValidateAnalysis(r); err == nil || !strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("expected unknown feature error, got %v", err)
	}
}

func TestValidateAnalysisMissingArtifacts(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}}}}}
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	if err := ValidateAnalysis(r); err == nil || !strings.Contains(err.Error(), "markdown missing") {
		t.Fatalf("expected missing markdown error, got %v", err)
	}
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis"), 0644) //nolint:errcheck
	if err := ValidateAnalysis(r); err == nil || !strings.Contains(err.Error(), "json missing") {
		t.Fatalf("expected missing json error, got %v", err)
	}
}
