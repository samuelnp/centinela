package workflow

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestValidateTestsErrorBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("tests/unit", 0755)                                                                                                                      //nolint:errcheck
	os.MkdirAll("tests/acceptance", 0755)                                                                                                                //nolint:errcheck
	os.WriteFile("tests/unit/x_test.go", []byte("x"), 0644)                                                                                              //nolint:errcheck
	os.WriteFile("tests/acceptance/x.go", []byte("package acceptance_test\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) { t.Log(\"ok\") }\n"), 0644) //nolint:errcheck

	err := validateTests("f", &config.Config{Workflow: config.WorkflowConfig{AcceptanceSuffix: ".steps.ts"}})
	if err == nil || !strings.Contains(err.Error(), "acceptance") {
		t.Fatalf("expected acceptance error, got %v", err)
	}

	err = validateTests("f", &config.Config{Validate: config.ValidateConfig{Commands: []string{"go test ./..."}}})
	if err == nil || !strings.Contains(err.Error(), "edge-case") {
		t.Fatalf("expected edge-case error, got %v", err)
	}

	if hasEdgeCaseReport("") {
		t.Fatal("expected empty feature edge-case check to be false")
	}
}
