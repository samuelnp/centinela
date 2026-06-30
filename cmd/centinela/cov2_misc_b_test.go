package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCov2HookPostwriteSkipsUnloadableJSON drives the continue branch when a
// .workflow/*.json file cannot be parsed as a workflow.
func TestCov2HookPostwriteSkipsUnloadableJSON(t *testing.T) {
	d := t.TempDir()
	if err := os.MkdirAll(filepath.Join(d, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, ".workflow", "broken.json"), []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	feedStdin(t, "", func() {
		if err := runHookPostwrite(nil, nil); err != nil {
			t.Fatalf("postwrite must skip unloadable JSON, got %v", err)
		}
	})
}

func TestCov2StartRejectsInvalidSlug(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := runStart(nil, []string{"Bad Slug!"}); err == nil {
		t.Fatal("expected an invalid-slug error")
	}
}

// TestCov2StartWorkflowDirCreateFails drives the MkdirAll error when .workflow
// already exists as a regular file.
func TestCov2StartWorkflowDirCreateFails(t *testing.T) {
	d := t.TempDir()
	t.Chdir(d)
	if err := os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".workflow", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runStart(nil, []string{"okfeat"}); err == nil || !strings.Contains(err.Error(), "cannot create") {
		t.Fatalf("expected a workflow-dir create error, got %v", err)
	}
}

// TestCov2PrecommitInstallSurfacesError forces githooks.Install's MkdirAll to
// fail by planting a regular file at .git (so .git/hooks cannot be created).
func TestCov2PrecommitInstallSurfacesError(t *testing.T) {
	d := t.TempDir()
	t.Chdir(d)
	if err := os.WriteFile(".git", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runPrecommitInstall(nil, nil); err == nil {
		t.Fatal("expected a hook install error")
	}
}

// TestCov2AuditSurfacesCorruptBaseline drives the audit.Load error branch via a
// configured baseline path that points at a malformed JSON file.
func TestCov2AuditSurfacesCorruptBaseline(t *testing.T) {
	d := t.TempDir()
	bp := filepath.Join(d, "baseline.json")
	if err := os.WriteFile(bp, []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[gates.audit_baseline]\nbaseline_path = \""+bp+"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	auditJSON = false
	if err := runAudit(nil, nil); err == nil {
		t.Fatal("expected a corrupt-baseline load error")
	}
}

// TestCov2HookCostNoActiveWorkflowNoOp: empty-cwd fallback + wf==nil no-op.
func TestCov2HookCostNoActiveWorkflowNoOp(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[cost]\nenabled=true\nstep_token_budget=1000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	tp := filepath.Join(d, "transcript.jsonl")
	if err := os.WriteFile(tp, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	feedStdin(t, `{"transcript_path":"`+tp+`"}`, func() {
		if err := runHookCost(nil, nil); err != nil {
			t.Fatalf("hook cost must be a silent no-op, got %v", err)
		}
	})
}
