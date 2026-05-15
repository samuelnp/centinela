package worktree_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

var allIgnoreFiles = []string{
	".gitignore", ".eslintignore", ".prettierignore",
	".dockerignore", ".rgignore",
}

func TestSyncIgnores_CreatesEntryInEveryIgnoreFile(t *testing.T) {
	repo := t.TempDir()
	res, err := worktree.SyncIgnores(repo)
	if err != nil {
		t.Fatalf("SyncIgnores: %v", err)
	}
	if len(res.Touched) < len(allIgnoreFiles) {
		t.Fatalf("expected at least %d files touched, got %d (%v)",
			len(allIgnoreFiles), len(res.Touched), res.Touched)
	}
	for _, name := range allIgnoreFiles {
		data, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(data), ".worktrees/") {
			t.Fatalf("%s missing `.worktrees/` entry: %q", name, data)
		}
	}
}

func TestSyncIgnores_IsIdempotent(t *testing.T) {
	repo := t.TempDir()
	if _, err := worktree.SyncIgnores(repo); err != nil {
		t.Fatalf("first SyncIgnores: %v", err)
	}
	snap := map[string][]byte{}
	for _, name := range allIgnoreFiles {
		snap[name], _ = os.ReadFile(filepath.Join(repo, name))
	}
	res, err := worktree.SyncIgnores(repo)
	if err != nil {
		t.Fatalf("second SyncIgnores: %v", err)
	}
	if len(res.Touched) != 0 {
		t.Fatalf("second run must touch nothing, got %v", res.Touched)
	}
	for _, name := range allIgnoreFiles {
		now, _ := os.ReadFile(filepath.Join(repo, name))
		if string(snap[name]) != string(now) {
			t.Fatalf("%s changed on second run: before=%q after=%q",
				name, snap[name], now)
		}
	}
}

func TestSyncIgnores_TsconfigMissingIsNoOp(t *testing.T) {
	repo := t.TempDir()
	if _, err := worktree.SyncIgnores(repo); err != nil {
		t.Fatalf("SyncIgnores: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "tsconfig.json")); !os.IsNotExist(err) {
		t.Fatalf("tsconfig.json should not be created when missing: err=%v", err)
	}
}
