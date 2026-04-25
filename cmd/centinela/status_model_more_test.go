package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunStatusModelFallsBackWithoutTTY(t *testing.T) {
	in, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	out, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	oldIn, oldOut := statusInput, statusOutput
	defer func() { statusInput, statusOutput = oldIn, oldOut }()
	statusInput, statusOutput = in, out
	wf := workflow.New("alpha")
	if err := runStatusModel([]*workflow.Workflow{wf}); err != nil {
		t.Fatalf("runStatusModel error: %v", err)
	}
	if _, err := out.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "Feature") || !strings.Contains(s, "alpha") {
		t.Fatalf("expected static status output, got %q", s)
	}
	if strings.Contains(s, "press any key to exit") {
		t.Fatalf("unexpected interactive hint in fallback output: %q", s)
	}
}

func TestHasTTYReturnsFalseForNilAndFiles(t *testing.T) {
	if hasTTY(nil) {
		t.Fatal("nil file should not be treated as tty")
	}
	f, err := os.CreateTemp(t.TempDir(), "file")
	if err != nil {
		t.Fatal(err)
	}
	if hasTTY(f) {
		t.Fatal("plain file should not be treated as tty")
	}
}

func TestRunStatusModelUsesInteractiveRunnerWhenTTY(t *testing.T) {
	oldTTY, oldRun := statusHasTTY, runInteractiveStatus
	defer func() { statusHasTTY, runInteractiveStatus = oldTTY, oldRun }()
	called := false
	statusHasTTY = func(*os.File) bool { return true }
	runInteractiveStatus = func(wfs []*workflow.Workflow) error {
		called = len(wfs) == 1 && wfs[0].Feature == "alpha"
		return nil
	}
	if err := runStatusModel([]*workflow.Workflow{workflow.New("alpha")}); err != nil {
		t.Fatalf("runStatusModel error: %v", err)
	}
	if !called {
		t.Fatal("expected interactive runner to be called")
	}
}
