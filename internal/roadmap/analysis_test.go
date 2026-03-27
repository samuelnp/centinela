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

func TestValidateAnalysisCycle(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}, {Name: "post"}}}}}
	os.MkdirAll(".workflow", 0755)                                    //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis"), 0644) //nolint:errcheck
	json := `{"role":"senior-product-manager","features":[{"name":"user","dependsOn":["post"]},{"name":"post","dependsOn":["user"]}]}`
	os.WriteFile(RoadmapAnalysisFile, []byte(json), 0644) //nolint:errcheck
	if err := ValidateAnalysis(r); err == nil || !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}
