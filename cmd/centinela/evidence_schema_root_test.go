package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// schemaWorktree lays out <tmp>/.worktrees/<feature> with workflow state and a
// brief, chdirs into it, and returns the checkout root.
func schemaWorktree(t *testing.T, feature, brief string) string {
	t.Helper()
	base := t.TempDir()
	if r, err := filepath.EvalSymlinks(base); err == nil {
		base = r
	}
	root := filepath.Join(base, ".worktrees", feature)
	for _, d := range []string{workflow.WorkflowDir, "docs/features", "internal/evidence"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
	if err := workflow.Save(workflow.New(feature)); err != nil {
		t.Fatal(err)
	}
	briefPath := filepath.Join("docs", "features", feature+".md")
	if err := os.WriteFile(briefPath, []byte(brief), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// schemaHandoff runs the command in dir and returns the printed feature and
// handoffTo.
func schemaHandoff(t *testing.T, dir, role string) (string, string) {
	t.Helper()
	t.Chdir(dir)
	out := captureStdout(t, func() {
		if err := runEvidenceSchema(nil, []string{role}); err != nil {
			t.Fatal(err)
		}
	})
	var got struct {
		Feature   string `json:"feature"`
		HandoffTo string `json:"handoffTo"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	return got.Feature, got.HandoffTo
}

// The headline guarantee, executable: the answer must not depend on how deep in
// the worktree the agent happens to stand. Before the fix the subdirectory run
// resolved the same correct slug but printed the legacy successor beside it,
// because the derivation's CWD-relative state lookup failed there.
func TestEvidenceSchemaSameFromRootAndSubdirectory(t *testing.T) {
	cases := []struct{ name, brief, role, want string }{
		{"internal feature skips the docs step", "# demo\n", "gatekeeper", "complete"},
		{"user-facing feature keeps the ux-ui hop", "surface: user-facing\n", "senior-engineer", "ux-ui-specialist"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := schemaWorktree(t, "demo", tc.brief)
			rf, rh := schemaHandoff(t, root, tc.role)
			sf, sh := schemaHandoff(t, filepath.Join(root, "internal", "evidence"), tc.role)
			if rf != "demo" || sf != "demo" {
				t.Fatalf("feature drifted: root %q, subdir %q", rf, sf)
			}
			if rh != tc.want || sh != tc.want {
				t.Fatalf("handoffTo root %q / subdir %q, want %q", rh, sh, tc.want)
			}
		})
	}
}

// A resolvable slug with no workflow state must answer the placeholder, not the
// legacy chain: `evidence init` refuses an unknown feature outright, so the
// schema command may not answer confidently where its own lookup failed.
func TestEvidenceSchemaStatelessWorktreeUsesPlaceholder(t *testing.T) {
	base := t.TempDir()
	deep := filepath.Join(base, ".worktrees", "not-a-feature", "deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	f, h := schemaHandoff(t, deep, "gatekeeper")
	if f != "<feature-slug>" || h != "<successor-role>" {
		t.Fatalf("fabricated worktree segment answered %q/%q", f, h)
	}
}
