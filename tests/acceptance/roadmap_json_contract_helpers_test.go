package acceptance_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	rmcOnce sync.Once
	rmcPath string
	rmcErr  string
)

// rmcBin builds the centinela binary once per test run from the repo root and
// returns its path. Never depends on the installed binary.
func rmcBin(t *testing.T) string {
	t.Helper()
	rmcOnce.Do(func() {
		o, _ := os.Getwd()
		repo := filepath.Clean(filepath.Join(o, "..", ".."))
		dir, err := os.MkdirTemp("", "rmcbin")
		if err != nil {
			rmcErr = err.Error()
			return
		}
		rmcPath = filepath.Join(dir, "centinela")
		b := exec.Command("go", "build", "-o", rmcPath, "./cmd/centinela")
		b.Dir = repo
		if out, err := b.CombinedOutput(); err != nil {
			rmcErr = err.Error() + "\n" + string(out)
		}
	})
	if rmcErr != "" {
		t.Fatalf("build centinela: %s", rmcErr)
	}
	return rmcPath
}

// rmcProject creates a temp project whose .workflow/roadmap.json holds body
// (skipped when body is empty, to exercise the missing-file paths).
func rmcProject(t *testing.T, body string) string {
	t.Helper()
	d := t.TempDir()
	if err := os.MkdirAll(filepath.Join(d, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if body != "" {
		p := filepath.Join(d, ".workflow", "roadmap.json")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

// rmcSeed writes a workflow-state file so FeatureStatus(feature) resolves to step.
func rmcSeed(t *testing.T, dir, feature, step string) {
	t.Helper()
	p := filepath.Join(dir, ".workflow", feature+".json")
	body := `{"feature":"` + feature + `","currentStep":"` + step + `","steps":{}}`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// rmcRun runs the binary in dir and returns stdout, stderr and the exit code.
func rmcRun(t *testing.T, dir string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(rmcBin(t), args...)
	cmd.Dir = dir
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	return so.String(), se.String(), code
}
