package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

const minifiedJSON = `{"feature":"alpha","step":"plan","role":"big-thinker","status":"done",` +
	`"generatedAt":"2026-05-12T00:00:00Z","inputs":[],"outputs":[],"edgeCases":[],"handoffTo":"feature-specialist"}`

func chdirWorktree(t *testing.T, feature string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), worktree.Dir, feature)
	if err := os.MkdirAll(filepath.Join(root, workflow.WorkflowDir), 0o755); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestHookPostwriteReformatsActiveFeatureJSON(t *testing.T) {
	root := chdirWorktree(t, "alpha")
	path := filepath.Join(root, workflow.WorkflowDir, "alpha-big-thinker.json")
	if err := os.WriteFile(path, []byte(minifiedJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	withStdin(t, `{"tool_input":{"file_path":"`+path+`"}}`, func() {
		if err := runHookPostwrite(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	data, _ := os.ReadFile(path)
	if !bytes.Contains(data, []byte("\n  \"feature\": \"alpha\"")) {
		t.Fatalf("file not reformatted:\n%s", data)
	}
}

func TestHookPostwriteIgnoresOtherFeatureFiles(t *testing.T) {
	root := chdirWorktree(t, "alpha")
	path := filepath.Join(root, workflow.WorkflowDir, "beta-big-thinker.json")
	if err := os.WriteFile(path, []byte(minifiedJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	withStdin(t, `{"tool_input":{"file_path":"`+path+`"}}`, func() {
		if err := runHookPostwrite(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	data, _ := os.ReadFile(path)
	if !bytes.Equal(data, []byte(minifiedJSON)) {
		t.Fatalf("beta file should be untouched, got: %s", data)
	}
}

func TestHookPostwriteLeavesNonJSONUntouched(t *testing.T) {
	root := chdirWorktree(t, "alpha")
	path := filepath.Join(root, workflow.WorkflowDir, "alpha-edge-cases.md")
	body := []byte("# not json\n")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	withStdin(t, `{"tool_input":{"file_path":"`+path+`"}}`, func() {
		if err := runHookPostwrite(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	data, _ := os.ReadFile(path)
	if !bytes.Equal(data, body) {
		t.Fatalf("md should be untouched, got: %s", data)
	}
}
