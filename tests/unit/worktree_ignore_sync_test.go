package unit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// allIgnoreFiles is the set SyncIgnores writes to (mirrors the implementation).
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
	// Snapshot file contents after first run.
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

func TestSyncIgnores_TsconfigPresentExcludePatched(t *testing.T) {
	repo := t.TempDir()
	cfg := `{"compilerOptions":{"target":"es2020"},"exclude":["node_modules"]}`
	if err := os.WriteFile(filepath.Join(repo, "tsconfig.json"), []byte(cfg), 0644); err != nil {
		t.Fatalf("write tsconfig: %v", err)
	}
	if _, err := worktree.SyncIgnores(repo); err != nil {
		t.Fatalf("SyncIgnores: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(repo, "tsconfig.json"))
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("post-patch tsconfig invalid JSON: %v: %s", err, raw)
	}
	var excl []string
	if err := json.Unmarshal(doc["exclude"], &excl); err != nil {
		t.Fatalf("exclude not array: %v", err)
	}
	hasNode, hasWt := false, false
	for _, e := range excl {
		if e == "node_modules" {
			hasNode = true
		}
		if e == ".worktrees" {
			hasWt = true
		}
	}
	if !hasWt || !hasNode {
		t.Fatalf("exclude lost or missing entries: %v", excl)
	}
}

func TestSyncIgnores_TsconfigMalformedJSON_NoOp(t *testing.T) {
	repo := t.TempDir()
	// tsconfig with a JS-style comment is not valid JSON: must be tolerated.
	cfg := "{ // comment\n  \"compilerOptions\": {} }"
	if err := os.WriteFile(filepath.Join(repo, "tsconfig.json"), []byte(cfg), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := worktree.SyncIgnores(repo); err != nil {
		t.Fatalf("SyncIgnores must tolerate malformed tsconfig: %v", err)
	}
	after, _ := os.ReadFile(filepath.Join(repo, "tsconfig.json"))
	if string(after) != cfg {
		t.Fatalf("malformed tsconfig was rewritten; expected no-op")
	}
}
