package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// projectCWD chdirs into a fresh temp project with a started workflow on the
// validate step and the given centinela.toml body.
func projectCWD(t *testing.T, toml string) {
	t.Helper()
	d := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(workflow.WorkflowDir, 0o755)            //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(toml), 0o644) //nolint:errcheck
	wf := workflow.New("feat")
	wf.CurrentStep = "validate"
	workflow.Save(wf) //nolint:errcheck
}

func parsePacket(t *testing.T, out string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, out)
	}
	return m
}

// Pass path: no gates enabled, no evidence → verdict pass, nil error, JSON on stdout.
func TestRunVerdict_PassEmitsJSON(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	projectCWD(t, "")
	var err error
	out := captureStdout(t, func() { err = runVerdict(nil, []string{"feat"}) })
	if err != nil {
		t.Fatalf("pass path must return nil, got %v", err)
	}
	m := parsePacket(t, out)
	if m["schema"] != "centinela.verdict/v1" {
		t.Fatalf("schema = %v", m["schema"])
	}
	if sum := m["summary"].(map[string]any); sum["verdict"] != "pass" {
		t.Fatalf("verdict = %v", sum["verdict"])
	}
}

// Fail path: an oversize source file under the file_size gate → sentinel error
// AND JSON still on stdout (the command writes before returning the error).
func TestRunVerdict_FailEmitsJSONAndSentinel(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	projectCWD(t, "[gates]\nfile_size = true\n")
	os.MkdirAll("internal", 0o755) //nolint:errcheck
	big := strings.Repeat("// line\n", 200)
	os.WriteFile(filepath.Join("internal", "big.go"), []byte("package p\n"+big), 0o644) //nolint:errcheck
	var err error
	out := captureStdout(t, func() { err = runVerdict(nil, []string{"feat"}) })
	if err == nil {
		t.Fatal("fail path must return a sentinel error")
	}
	m := parsePacket(t, out)
	if sum := m["summary"].(map[string]any); sum["verdict"] != "fail" || sum["exitCode"].(float64) != 1 {
		t.Fatalf("fail summary wrong: %v", sum)
	}
}

// The resolved headless state surfaces on the packet run.headless field.
func TestRunVerdict_HeadlessFlag(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	projectCWD(t, "")
	var err error
	out := captureStdout(t, func() { err = runVerdict(nil, []string{"feat"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	run := parsePacket(t, out)["run"].(map[string]any)
	if run["headless"] != true {
		t.Fatalf("run.headless = %v, want true", run["headless"])
	}
}
