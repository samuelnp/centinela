package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookContextWithWorkflowAndRoadmap(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll(".workflow", 0755)                                                                               //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"f"}]}]}`), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "plan"
	workflow.Save(wf)                                           //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                          //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                             //nolint:errcheck
	os.MkdirAll("specs", 0755)                                  //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("x"), 0644)       //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("x"), 0644)          //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x"), 0644) //nolint:errcheck

	withStdin(t, "{}", func() {
		if err := runHookContext(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
}
