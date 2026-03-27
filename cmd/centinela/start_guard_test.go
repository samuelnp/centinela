package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestWorkflowOrderForFeatureExistingProject(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                     //nolint:errcheck
	os.Chdir(d)                                                           //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	order, err := workflowOrderForFeature("feature-x")
	if err != nil || len(order) != 4 || order[2] != "tests" {
		t.Fatalf("expected default order for existing project: %v %v", order, err)
	}
}

func TestWorkflowOrderForFeatureGreenfieldBlocksNonBootstrap(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                       //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "setup"}}}}}
	roadmap.Save(r) //nolint:errcheck
	writeRoadmapAnalysis(t, "setup")
	if _, err := workflowOrderForFeature("feature-x"); err == nil {
		t.Fatal("expected greenfield non-bootstrap to be blocked")
	}
}

func TestWorkflowOrderForFeatureGreenfieldBootstrapUsesThreeSteps(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                       //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "setup"}}}}}
	roadmap.Save(r) //nolint:errcheck
	writeRoadmapAnalysis(t, "setup")
	order, err := workflowOrderForFeature("setup")
	if err != nil || len(order) != 3 || order[2] != "validate" {
		t.Fatalf("expected bootstrap order: %v %v", order, err)
	}
	wf := workflow.NewWithOrder("setup", order)
	if workflow.StepNumberFor(wf, "validate") != 3 {
		t.Fatal("validate should be step 3 for bootstrap workflow")
	}
}

func TestWorkflowOrderForFeatureGreenfieldRequiresRoadmapAndBootstrapPhase(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                       //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	if _, err := workflowOrderForFeature("x"); err == nil {
		t.Fatal("expected error when roadmap is missing")
	}
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 1", Features: []roadmap.Feature{{Name: "x"}}}}}
	roadmap.Save(r) //nolint:errcheck
	writeRoadmapAnalysis(t, "x")
	if _, err := workflowOrderForFeature("x"); err == nil {
		t.Fatal("expected error when bootstrap phase is missing")
	}
}

func TestWorkflowOrderForFeatureGreenfieldAllowsAfterBootstrapComplete(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                       //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "setup"}}}, {Name: "Phase 1", Features: []roadmap.Feature{{Name: "feature-x"}}}}}
	roadmap.Save(r) //nolint:errcheck
	writeRoadmapAnalysis(t, "setup", "feature-x")
	os.MkdirAll(workflow.WorkflowDir, 0755)                                                                          //nolint:errcheck
	workflow.Save(&workflow.Workflow{Feature: "setup", CurrentStep: "done", Steps: map[string]workflow.StepState{}}) //nolint:errcheck
	order, err := workflowOrderForFeature("feature-x")
	if err != nil || len(order) != 4 || order[2] != "tests" {
		t.Fatalf("expected default order after bootstrap complete: %v %v", order, err)
	}
}
