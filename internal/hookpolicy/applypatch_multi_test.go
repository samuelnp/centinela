package hookpolicy

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func codeCfg() *config.Config {
	cfg := &config.Config{}
	cfg.Workflow.CodeDirs = []string{"/internal/"}
	return cfg
}

// TestEvaluatePrewriteMulti_RelativeCodeBlocks is the regression guard: a
// repo-RELATIVE apply_patch code path with no active workflow MUST block. The
// original bug passed the relative path straight through, so it never blocked.
func TestEvaluatePrewriteMulti_RelativeCodeBlocks(t *testing.T) {
	cwd := t.TempDir()
	d := EvaluatePrewriteMulti([]string{"internal/foo.go"}, cwd, codeCfg(), nil)
	if d.Allow {
		t.Fatal("relative code path with no workflow must NOT be allowed")
	}
	if !d.NeedInit {
		t.Fatalf("expected NeedInit, got %+v", d)
	}
	if d.Path != "internal/foo.go" {
		t.Fatalf("Path should be the original relative path, got %q", d.Path)
	}
}

func TestEvaluatePrewriteMulti_RelativeDocsAllowed(t *testing.T) {
	cwd := t.TempDir()
	if !EvaluatePrewriteMulti([]string{"docs/notes.md"}, cwd, codeCfg(), nil).Allow {
		t.Fatal("relative docs path should be allowed")
	}
}

func TestEvaluatePrewriteMulti_AbsoluteBlocks(t *testing.T) {
	cwd := t.TempDir()
	d := EvaluatePrewriteMulti([]string{cwd + "/internal/foo.go"}, cwd, codeCfg(), nil)
	if d.Allow || !d.NeedInit {
		t.Fatalf("absolute code path under cwd should block, got %+v", d)
	}
}

func TestEvaluatePrewriteMulti_FirstBlockingWins(t *testing.T) {
	cwd := t.TempDir()
	paths := []string{"docs/ok.md", "internal/bad.go", "internal/also.go"}
	d := EvaluatePrewriteMulti(paths, cwd, codeCfg(), nil)
	if d.Allow || d.Path != "internal/bad.go" {
		t.Fatalf("first blocking path should win, got %+v", d)
	}
}

func TestEvaluatePrewriteMulti_EmptyAndAllAllowed(t *testing.T) {
	cwd := t.TempDir()
	if !EvaluatePrewriteMulti(nil, cwd, codeCfg(), nil).Allow {
		t.Fatal("no paths should be allowed (no-op)")
	}
	wfs := []*workflow.Workflow{{Feature: "f", CurrentStep: "code"}}
	if !EvaluatePrewriteMulti([]string{"internal/foo.go"}, cwd, codeCfg(), wfs).Allow {
		t.Fatal("code path during code step should be allowed")
	}
}
