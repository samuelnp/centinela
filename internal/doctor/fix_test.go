package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixRepairsSafeAndPreservesDestructive(t *testing.T) {
	dir := repoFixture(t)
	// fixable: missing hooks, drifted roadmap, orphaned tmp.
	writeFile(t, ".claude/settings.json", "{}")
	seedRoadmap(t, "Phase 1: Core")
	writeFile(t, "ROADMAP.md", "stale\n")
	writeFile(t, ".workflow/feat-qa-senior.json.tmp", "{}")
	// destructive: abandoned worktree.
	_ = os.MkdirAll(filepath.Join(dir, ".worktrees", "gone"), 0o755)
	seedMakefile(t, dir, "0.21.1")
	stubVersion(t, func() (string, error) { return "0.21.1\n", nil })
	stubGit(t, func(repo string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" && len(args) > 1 && args[1] == "--is-inside-work-tree" {
			return []byte("true\n"), nil
		}
		return nil, errStub // every branch missing => abandoned
	})
	ctx := Context{Root: dir, Config: configWithTimeout(240)}
	post := Fix(ctx)
	byName := map[string]Diagnosis{}
	for _, d := range post {
		byName[d.Name] = d
	}
	for _, n := range []string{"hooks", "roadmap", "evidence"} {
		if byName[n].Status != OK {
			t.Fatalf("%s must be OK post-fix, got %v", n, byName[n].Status)
		}
	}
	// destructive untouched.
	if byName["worktrees"].Status == OK {
		t.Fatal("abandoned worktree must remain reported, not OK")
	}
	if !dirExists(filepath.Join(dir, ".worktrees", "gone")) {
		t.Fatal("--fix must never remove a worktree")
	}
	if left, _ := filepath.Glob(".workflow/*.json.tmp"); len(left) != 0 {
		t.Fatalf("tmp files should be swept, left %v", left)
	}
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func TestFixPartialFailureMarksCheckError(t *testing.T) {
	dir := repoFixture(t)
	seedRoadmap(t, "✅ Phase 0: Bootstrap") // glyph => fixable, but...
	// make roadmap.json unwritable so the safe repair's Save fails.
	if err := os.Chmod(".workflow/roadmap.json", 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(".workflow", 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(".workflow", 0o755) })
	// also a working fixable repair to prove others still run.
	writeFile(t, ".claude/settings.json", "{}")
	seedMakefile(t, dir, "0.21.1")
	stubVersion(t, func() (string, error) { return "0.21.1\n", nil })
	stubGit(t, okGit(""))
	ctx := Context{Root: dir, Config: configWithTimeout(240)}
	post := Fix(ctx)
	byName := map[string]Diagnosis{}
	for _, d := range post {
		byName[d.Name] = d
	}
	if byName["roadmap"].Status != Error {
		t.Fatalf("failing repair must surface as Error, got %v", byName["roadmap"].Status)
	}
	if byName["hooks"].Status != OK {
		t.Fatalf("other repairs must still run, hooks=%v", byName["hooks"].Status)
	}
	if !ExitError(post) {
		t.Fatal("a failed repair must drive exit error")
	}
}
