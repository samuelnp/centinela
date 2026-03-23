package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestValidateArtifactsPlanAndGatekeeper(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("docs/plans", 0755)                             //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("# plan"), 0644)     //nolint:errcheck
	os.MkdirAll("specs", 0755)                                  //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x"), 0644) //nolint:errcheck
	if err := ValidateArtifacts("f", "plan", &config.Config{}); err != nil {
		t.Fatalf("plan should pass: %v", err)
	}
	if err := ValidateArtifacts("f", "validate", &config.Config{}); err == nil {
		t.Fatal("expected missing gatekeeper error")
	}
	os.MkdirAll(".workflow", 0755)                                  //nolint:errcheck
	os.WriteFile(".workflow/f-gatekeeper.md", []byte("SAFE"), 0644) //nolint:errcheck
	if err := ValidateArtifacts("f", "validate", &config.Config{}); err != nil {
		t.Fatalf("validate should pass with gatekeeper: %v", err)
	}
}

func TestProductionReadinessBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	cfg := &config.Config{Gates: config.GatesConfig{ProductionReadinessEnabled: true}}
	if err := validateProductionReadiness("f", cfg); err == nil {
		t.Fatal("expected missing production readiness report")
	}
	os.MkdirAll(".workflow", 0755)                                                           //nolint:errcheck
	os.WriteFile(".workflow/f-production-readiness.md", []byte("**Status:** WARNING"), 0644) //nolint:errcheck
	if err := validateProductionReadiness("f", cfg); err != nil {
		t.Fatalf("warning report should pass: %v", err)
	}
	if w := ProductionReadinessWarning("f", cfg); w != "f" {
		t.Fatalf("expected warning for feature, got %q", w)
	}
	os.WriteFile(".workflow/f-production-readiness.md", []byte("**Status:** BLOCKING"), 0644) //nolint:errcheck
	if err := checkPRStatus("**Status:** BLOCKING", "f"); err == nil {
		t.Fatal("expected blocking status error")
	}
}

func TestValidatePlanAndWarningBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("docs/plans", 0755)                    //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("x"), 0644) //nolint:errcheck
	if err := validatePlan("f"); err == nil {
		t.Fatal("expected missing specs error")
	}
	os.Remove("docs/plans/f.md") //nolint:errcheck
	if err := validatePlan("f"); err == nil {
		t.Fatal("expected missing plan file error")
	}
	os.WriteFile("docs/plans/f.md", []byte("x"), 0644)          //nolint:errcheck
	os.MkdirAll("specs", 0755)                                  //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x"), 0644) //nolint:errcheck
	if err := validatePlan("f"); err != nil {
		t.Fatalf("expected validatePlan success, got %v", err)
	}
	if w := ProductionReadinessWarning("f", &config.Config{}); w != "" {
		t.Fatal("warning should be empty when gate disabled")
	}
	cfg := &config.Config{Gates: config.GatesConfig{ProductionReadinessEnabled: true}}
	if w := ProductionReadinessWarning("f", cfg); w != "" {
		t.Fatal("warning should be empty when report missing")
	}
	os.MkdirAll(".workflow", 0755)                                            //nolint:errcheck
	os.WriteFile(".workflow/f-production-readiness.md", []byte("SAFE"), 0644) //nolint:errcheck
	if w := ProductionReadinessWarning("f", cfg); w != "" {
		t.Fatal("warning should be empty for SAFE report")
	}
}
