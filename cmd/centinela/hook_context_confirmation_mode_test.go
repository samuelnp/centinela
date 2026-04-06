package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookContextReviewPromptModes(t *testing.T) {
	t.Run("every_step prompts", func(t *testing.T) {
		out := runContextWithMode(t, "", "code")
		if !strings.Contains(out, "shall I advance") {
			t.Fatalf("expected review prompt, got: %s", out)
		}
	})
	t.Run("after_plan prompts only plan", func(t *testing.T) {
		if out := runContextWithMode(t, "after_plan", "code"); strings.Contains(out, "shall I advance") {
			t.Fatalf("did not expect review prompt for code: %s", out)
		}
		if out := runContextWithMode(t, "after_plan", "plan"); !strings.Contains(out, "shall I advance") {
			t.Fatalf("expected review prompt for plan: %s", out)
		}
	})
	t.Run("auto suppresses prompts", func(t *testing.T) {
		out := runContextWithMode(t, "auto", "plan")
		if strings.Contains(out, "shall I advance") {
			t.Fatalf("did not expect review prompt, got: %s", out)
		}
	})
}

func runContextWithMode(t *testing.T, mode, step string) string {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if mode == "" {
		os.WriteFile("centinela.toml", []byte("[workflow]\ndisable_auto_commit=true\n"), 0644) //nolint:errcheck
	} else {
		c := "[workflow]\ndisable_auto_commit=true\nstep_confirmation_mode=\"" + mode + "\"\n"
		os.WriteFile("centinela.toml", []byte(c), 0644) //nolint:errcheck
	}
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = step
	wf.OrchestrationMode = ""
	workflow.Save(wf)                                           //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                          //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                             //nolint:errcheck
	os.MkdirAll("specs", 0755)                                  //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("x"), 0644)       //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("x"), 0644)          //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x"), 0644) //nolint:errcheck
	return captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookContext(nil, nil) }) //nolint:errcheck
	})
}
