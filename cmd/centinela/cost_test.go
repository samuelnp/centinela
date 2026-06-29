package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// seedCostRepo builds a temp repo with [cost] enabled, an active workflow, and a
// telemetry log holding one cost sample for demo/code (2500 tokens). Returns the
// dir and chdirs into it for the test.
func seedCostRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	must := func(p, body string) {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must(filepath.Join(dir, "centinela.toml"),
		"[cost]\nenabled=true\nstep_token_budget=1000\nfeature_token_budget=5000\n")
	must(filepath.Join(dir, ".workflow", "demo.json"),
		`{"feature":"demo","currentStep":"code","stepOrder":["plan","code"],"steps":{}}`)
	must(filepath.Join(dir, ".workflow", "telemetry", "events.jsonl"),
		`{"schema":"centinela.telemetry/v1","type":"cost-sample","feature":"demo","step":"code","inputTokens":1500,"outputTokens":1000}`+"\n")
	t.Chdir(dir)
	return dir
}

func capture(t *testing.T, fn func() error) string {
	t.Helper()
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	err := fn()
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func TestActiveWorkflowRootMode(t *testing.T) {
	seedCostRepo(t)
	wf := activeWorkflow(mustGetwd())
	if wf == nil || wf.Feature != "demo" || wf.CurrentStep != "code" {
		t.Fatalf("expected demo/code active workflow, got %+v", wf)
	}
}

func TestRunCostShowsOverBudget(t *testing.T) {
	seedCostRepo(t)
	costJSON = false
	out := capture(t, func() error { return runCost(nil, nil) })
	if !strings.Contains(out, "demo/code") || !strings.Contains(out, "OVER") {
		t.Fatalf("expected over-budget step in report, got %q", out)
	}
}

func TestEmitCostWarningOverBudget(t *testing.T) {
	seedCostRepo(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	out := capture(t, func() error { emitCostWarning(cfg); return nil })
	if !strings.Contains(out, "over budget") {
		t.Fatalf("expected soft-gate warning, got %q", out)
	}
}

func TestEmitCostWarningInactiveIsSilent(t *testing.T) {
	seedCostRepo(t)
	out := capture(t, func() error { emitCostWarning(&config.Config{}); return nil })
	if strings.TrimSpace(out) != "" {
		t.Fatalf("inactive cost should be silent, got %q", out)
	}
}
