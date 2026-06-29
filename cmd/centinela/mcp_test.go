package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// seedMcpRepo chdirs into a temp repo with file_size-only gates and one active
// workflow at the plan step.
func seedMcpRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	w := func(rel, body string) {
		p := filepath.Join(dir, rel)
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w("centinela.toml", "[gates]\nfile_size = true\n[i18n]\nlocales = [\"en\", \"es\"]\n")
	w(".workflow/demo.json", `{"feature":"demo","currentStep":"plan","stepOrder":["plan","code"],"steps":{}}`)
	t.Chdir(dir)
	return dir
}

func TestMcpRulesSurface(t *testing.T) {
	seedMcpRepo(t)
	r := mcpRules()
	if r.MaxFileLines != 100 || len(r.Locales) != 2 {
		t.Fatalf("unexpected rules: %+v", r)
	}
	found := false
	for _, g := range r.Gates {
		if g == "file_size" {
			found = true
		}
	}
	if !found {
		t.Fatalf("file_size gate not reported: %+v", r.Gates)
	}
}

func TestEnabledGateNames(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.FileSizeEnabled = true
	cfg.Gates.Security.Enabled = true
	names := enabledGateNames(cfg)
	if len(names) != 2 || names[0] != "file_size" || names[1] != "security" {
		t.Fatalf("unexpected gate names: %v", names)
	}
}

func TestWorkflowForFeature(t *testing.T) {
	seedMcpRepo(t)
	if wf := workflowForFeature("demo"); wf == nil || wf.Feature != "demo" {
		t.Fatalf("named lookup failed: %+v", wf)
	}
	if wf := workflowForFeature(""); wf == nil || wf.Feature != "demo" {
		t.Fatalf("active lookup failed: %+v", wf)
	}
	if wf := workflowForFeature("nope"); wf != nil {
		t.Fatalf("unknown feature should be nil, got %+v", wf)
	}
}

func TestMcpVerdictAssembles(t *testing.T) {
	seedMcpRepo(t)
	p, err := mcpVerdict("demo")
	if err != nil || p == nil || p.Run.Feature != "demo" {
		t.Fatalf("mcpVerdict: %+v err=%v", p, err)
	}
	if p.Schema != "centinela.verdict/v1" {
		t.Fatalf("unexpected packet schema: %s", p.Schema)
	}
}

func TestMcpVerdictNoActiveFeature(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := mcpVerdict(""); err == nil {
		t.Fatal("expected error when no active feature")
	}
}
