package roadmap

import (
	"os"
	"strings"
	"testing"
)

func TestValidateQualityPassAndThreshold(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}, {Name: "post"}}}}}
	os.MkdirAll(".workflow", 0755)                                  //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality"), 0644) //nolint:errcheck
	ok := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"user","scores":{"acceptanceCriteria":9,"userValue":10,"definitionClarity":9,"dependencies":9,"effortEstimation":2,"overall":9},"summary":"ok"},{"name":"post","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":3,"overall":10},"summary":"ok"}]}`
	os.WriteFile(RoadmapQualityFile, []byte(ok), 0644) //nolint:errcheck
	if err := ValidateQuality(r); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
	bad := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"user","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":2,"overall":8},"summary":"low"},{"name":"post","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":3,"overall":10},"summary":"ok"}]}`
	os.WriteFile(RoadmapQualityFile, []byte(bad), 0644) //nolint:errcheck
	if err := ValidateQuality(r); err == nil || !strings.Contains(err.Error(), "below 9") {
		t.Fatalf("expected threshold error, got %v", err)
	}
}

func TestValidateQualityInvalidBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}}}}}
	os.MkdirAll(".workflow", 0755)                                  //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality"), 0644) //nolint:errcheck
	os.WriteFile(RoadmapQualityFile, []byte(`{bad`), 0644)          //nolint:errcheck
	if err := ValidateQuality(r); err == nil || !strings.Contains(err.Error(), "invalid roadmap quality") {
		t.Fatalf("expected invalid json error, got %v", err)
	}
	role := `{"role":"qa","threshold":9,"features":[{"name":"user","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":2,"overall":9},"summary":"ok"}]}`
	os.WriteFile(RoadmapQualityFile, []byte(role), 0644) //nolint:errcheck
	if err := ValidateQuality(r); err == nil || !strings.Contains(err.Error(), "role") {
		t.Fatalf("expected role error, got %v", err)
	}
}
