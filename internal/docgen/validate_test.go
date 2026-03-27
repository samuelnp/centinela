package docgen

import (
	"os"
	"strings"
	"testing"
)

func TestValidateInputs(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if err := ValidateInputs(); err == nil || !strings.Contains(err.Error(), "PROJECT.md") {
		t.Fatalf("expected project error, got %v", err)
	}
	writeFixture(t)
	os.Remove(".workflow/roadmap-analysis.md")   //nolint:errcheck
	os.Remove(".workflow/roadmap-analysis.json") //nolint:errcheck
	if err := ValidateInputs(); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}
