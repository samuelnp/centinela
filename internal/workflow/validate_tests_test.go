package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestValidateTests_DefaultAndSuffixes(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("tests/unit", 0755)                          //nolint:errcheck
	os.MkdirAll("tests/acceptance", 0755)                    //nolint:errcheck
	os.WriteFile("tests/unit/a.go", []byte("x"), 0644)       //nolint:errcheck
	os.WriteFile("tests/acceptance/a.go", []byte("x"), 0644) //nolint:errcheck
	if err := validateTests(&config.Config{}); err != nil {
		t.Fatalf("default validateTests failed: %v", err)
	}
	if err := validateTests(&config.Config{Workflow: config.WorkflowConfig{TestSuffixes: []string{"_test.go"}, AcceptanceSuffix: ".steps.ts"}}); err == nil {
		t.Fatal("expected failure for missing suffix-matching files")
	}
}

func TestHasHelpers(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("tests/integration", 0755)                         //nolint:errcheck
	os.WriteFile("tests/integration/x_test.go", []byte("x"), 0644) //nolint:errcheck
	if !hasAnyFile("tests/integration") || !hasFileSuffix("tests/integration", "_test.go") {
		t.Fatal("expected helper detection true")
	}
	if hasFileSuffix("tests/integration", ".steps.ts") {
		t.Fatal("unexpected suffix match")
	}
}
