package roadmap

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestBootstrapHelpers(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "Phase 0: Bootstrap", Features: []Feature{{Name: "setup"}}}}}
	if !HasBootstrapPhase(r) || !IsBootstrapFeature(r, "setup") {
		t.Fatal("expected bootstrap phase detection")
	}
	if IsBootstrapFeature(r, "other") {
		t.Fatal("unexpected bootstrap feature detection")
	}
}

func TestBootstrapComplete(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Name: "Phase 0: Bootstrap", Features: []Feature{{Name: "setup"}}}}}
	if BootstrapComplete(r) {
		t.Fatal("bootstrap cannot be complete without workflow")
	}
	workflow.Save(&workflow.Workflow{Feature: "setup", CurrentStep: "done", Steps: map[string]workflow.StepState{}}) //nolint:errcheck
	if !BootstrapComplete(r) {
		t.Fatal("bootstrap should be complete when setup workflow is done")
	}
}
