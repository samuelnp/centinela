package unit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestTestsStep_RequiresEdgeCaseReport(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("tests/unit", 0755)                               //nolint:errcheck
	os.MkdirAll("tests/acceptance", 0755)                         //nolint:errcheck
	os.WriteFile("tests/unit/a_test.go", []byte("x"), 0644)       //nolint:errcheck
	os.WriteFile("tests/acceptance/a_test.go", []byte("x"), 0644) //nolint:errcheck

	if err := workflow.ValidateArtifacts("f", "tests", &config.Config{}); err == nil {
		t.Fatal("expected failure when edge-case report is missing")
	}
	os.MkdirAll(filepath.Join(".workflow"), 0755)                 //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("ok"), 0644) //nolint:errcheck
	if err := workflow.ValidateArtifacts("f", "tests", &config.Config{}); err != nil {
		t.Fatalf("expected tests step to pass with edge-case report: %v", err)
	}
}
