package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func withStdin(t *testing.T, content string, fn func()) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content) //nolint:errcheck
	f.Seek(0, 0)           //nolint:errcheck
	old := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = old }()
	fn()
}

func TestHookContextAndSetupNoWorkflows(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	withStdin(t, "{}", func() {
		if err := runHookContext(nil, nil); err != nil {
			t.Fatalf("runHookContext: %v", err)
		}
		if err := runHookSetup(nil, nil); err != nil {
			t.Fatalf("runHookSetup: %v", err)
		}
	})
}

func TestHookPostwriteAndPrewriteNoopPaths(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	workflow.Save(workflow.New("f"))        //nolint:errcheck

	withStdin(t, "{}", func() {
		if err := runHookPostwrite(nil, nil); err != nil {
			t.Fatalf("runHookPostwrite: %v", err)
		}
		if err := runHookPrewrite(nil, nil); err != nil {
			t.Fatalf("runHookPrewrite: %v", err)
		}
	})
}
