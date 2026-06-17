package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func realTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		return r
	}
	return dir
}

func TestResolveRootNonWorktree(t *testing.T) {
	dir := realTemp(t)
	if got := resolveRoot(dir); got != dir {
		t.Fatalf("non-worktree root must be cwd, got %q want %q", got, dir)
	}
}

func TestResolveRootClimbsOutOfWorktree(t *testing.T) {
	root := realTemp(t)
	wt := filepath.Join(root, ".worktrees", "feat")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := resolveRoot(wt); got != root {
		t.Fatalf("worktree must climb to parent of .worktrees, got %q want %q", got, root)
	}
}

func TestNewContextChdirsToRootAndLoadsConfig(t *testing.T) {
	root := realTemp(t)
	if err := os.WriteFile(filepath.Join(root, "centinela.toml"),
		[]byte("[verify]\nverify_timeout = 240\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	ctx, err := NewContext(root)
	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	if ctx.Root != root || ctx.Config == nil || ctx.CfgErr != nil {
		t.Fatalf("expected loaded config at root, got root=%q cfg=%v err=%v", ctx.Root, ctx.Config, ctx.CfgErr)
	}
	cwd, _ := os.Getwd()
	if cwd != root {
		t.Fatalf("NewContext must chdir to root, cwd=%q", cwd)
	}
}

func TestNewContextCapturesParseError(t *testing.T) {
	root := realTemp(t)
	if err := os.WriteFile(filepath.Join(root, "centinela.toml"), []byte("[bad\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	ctx, err := NewContext(root)
	if err != nil {
		t.Fatalf("parse error must be captured, not returned: %v", err)
	}
	if ctx.CfgErr == nil || ctx.Config != nil {
		t.Fatalf("parse error must populate CfgErr, got cfg=%v err=%v", ctx.Config, ctx.CfgErr)
	}
}
