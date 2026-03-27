package docgen

import (
	"os"
	"strings"
	"testing"
)

func TestValidateInputsBranchCoverage(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                               //nolint:errcheck
	os.Chdir(d)                                     //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("# P"), 0644) //nolint:errcheck
	if err := ValidateInputs(); err == nil || !strings.Contains(err.Error(), "ROADMAP.md") {
		t.Fatalf("expected missing ROADMAP error, got %v", err)
	}
	os.WriteFile("ROADMAP.md", []byte("# R"), 0644) //nolint:errcheck
	if err := ValidateInputs(); err == nil || !strings.Contains(err.Error(), "roadmap.json") {
		t.Fatalf("expected missing roadmap json error, got %v", err)
	}
	os.MkdirAll(".workflow", 0755)                                                                              //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P","features":[{"name":"f"}]}]}`), 0644) //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.md", []byte("# A"), 0644)                                          //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{"features":[]}`), 0644)                            //nolint:errcheck
	if err := ValidateInputs(); err == nil || !strings.Contains(err.Error(), "invalid roadmap analysis") {
		t.Fatalf("expected invalid analysis error, got %v", err)
	}
}
