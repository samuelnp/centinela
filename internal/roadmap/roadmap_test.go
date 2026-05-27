package roadmap

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	r := &Roadmap{Phases: []Phase{{Name: "P1", Features: []Feature{{Name: "f1"}}}}}
	if err := Save(r); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil || len(got.Phases) != 1 {
		t.Fatalf("Load: %v %#v", err, got)
	}
}

func TestLoadAndSaveErrors(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(".workflow", 0755)                  //nolint:errcheck
	os.WriteFile(RoadmapFile, []byte("{bad"), 0644) //nolint:errcheck
	if _, err := Load(); err == nil {
		t.Fatal("expected json parse error")
	}
	os.Remove(RoadmapFile) //nolint:errcheck
	if _, err := Load(); err == nil {
		t.Fatal("expected missing roadmap file error")
	}
	os.RemoveAll(".workflow")                    //nolint:errcheck
	os.WriteFile(".workflow", []byte("x"), 0644) //nolint:errcheck
	if err := Save(&Roadmap{}); err == nil {
		t.Fatal("expected mkdir conflict error")
	}
	os.Remove(".workflow")                      //nolint:errcheck
	os.MkdirAll(".workflow/roadmap.json", 0755) //nolint:errcheck
	if err := Save(&Roadmap{}); err == nil {
		t.Fatal("expected write error when roadmap path is directory")
	}
}

func TestFeatureStatusAndSummary(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(workflow.WorkflowDir, 0755)                                                                          //nolint:errcheck
	workflow.Save(&workflow.Workflow{Feature: "done", CurrentStep: "done", Steps: map[string]workflow.StepState{}})  //nolint:errcheck
	workflow.Save(&workflow.Workflow{Feature: "doing", CurrentStep: "code", Steps: map[string]workflow.StepState{}}) //nolint:errcheck
	r := &Roadmap{Phases: []Phase{{Features: []Feature{{Name: "done"}, {Name: "doing"}, {Name: "new"}}}}}
	p, ip, dn := r.Summary()
	if p != 1 || ip != 1 || dn != 1 {
		t.Fatalf("unexpected summary: %d %d %d", p, ip, dn)
	}
}
