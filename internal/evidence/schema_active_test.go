package evidence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// worktreeCwd builds <tmp>/.worktrees/<feature>/<rel...> and returns the
// checkout root and the deepest directory, both symlink-resolved so they can be
// compared with what DetectFeature returns.
func worktreeCwd(t *testing.T, feature string, rel ...string) (root, deep string) {
	t.Helper()
	base := t.TempDir()
	if r, err := filepath.EvalSymlinks(base); err == nil {
		base = r
	}
	root = filepath.Join(base, ".worktrees", feature)
	deep = filepath.Join(append([]string{root}, rel...)...)
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	return root, deep
}

// The worktree signal must answer identically from the checkout root and from
// any depth inside it — the resolver half of root-vs-subdirectory equivalence.
// Returning the root is what lets the caller derive where the state lives.
func TestResolveActiveFeatureWorktreeAnyDepth(t *testing.T) {
	root, deep := worktreeCwd(t, "demo-wt", "internal", "evidence")
	for _, cwd := range []string{root, deep} {
		f, r := ResolveActiveFeature(cwd)
		if f != "demo-wt" || r != root {
			t.Fatalf("ResolveActiveFeature(%q) = %q/%q, want demo-wt/%q", cwd, f, r, root)
		}
	}
}

// A worktree names exactly one feature, so it outranks the ambient scan even
// when a different workflow is the only active one in the CWD's .workflow.
func TestResolveActiveFeatureWorktreeBeatsActiveScan(t *testing.T) {
	chdirToTemp(t)
	if err := workflow.Save(workflow.New("unrelated")); err != nil {
		t.Fatal(err)
	}
	_, deep := worktreeCwd(t, "demo-wt", "sub")
	if f, _ := ResolveActiveFeature(deep); f != "demo-wt" {
		t.Fatalf("worktree signal lost to the active scan: %q", f)
	}
}

// Outside a worktree the scan may answer only when it is UNAMBIGUOUS: exactly
// one active workflow. Zero and two-or-more are both "no answer" — never the
// most-recently-touched of several parallel sessions.
func TestResolveActiveFeatureAmbientScan(t *testing.T) {
	cases := []struct {
		name     string
		features []string
		want     string
	}{
		{"zero active", nil, ""},
		{"exactly one active", []string{"demo-solo"}, "demo-solo"},
		{"two active never guesses", []string{"demo-a", "demo-b"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chdirToTemp(t)
			for _, f := range tc.features {
				if err := workflow.Save(workflow.New(f)); err != nil {
					t.Fatal(err)
				}
			}
			f, r := ResolveActiveFeature(t.TempDir())
			if f != tc.want || r != "" {
				t.Fatalf("ResolveActiveFeature = %q/%q, want %q with no root", f, r, tc.want)
			}
		})
	}
}

// E5, asserted rather than assumed: ActiveWorkflows warns on an unparseable
// sibling and skips it, so a corrupt file cannot turn "one active" into "none".
// It CAN turn "two active" into a confident one — deferred as
// evidence-active-workflow-corrupt-json-ambiguity, pinned here so the fix has a
// failing test to flip.
func TestResolveActiveFeatureCorruptSiblingIsSkipped(t *testing.T) {
	chdirToTemp(t)
	if err := workflow.Save(workflow.New("demo-active")); err != nil {
		t.Fatal(err)
	}
	corrupt := filepath.Join(workflow.WorkflowDir, "demo-corrupt.json")
	if err := os.WriteFile(corrupt, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if f, _ := ResolveActiveFeature(t.TempDir()); f != "demo-active" {
		t.Fatalf("corrupt sibling changed the answer: %q", f)
	}
}
