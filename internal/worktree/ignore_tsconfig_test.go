package worktree_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestSyncIgnores_TsconfigMissingExcludeKey(t *testing.T) {
	repo := t.TempDir()
	cfg := `{"compilerOptions":{"strict":true}}`
	if err := os.WriteFile(filepath.Join(repo, "tsconfig.json"), []byte(cfg), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := worktree.SyncIgnores(repo); err != nil {
		t.Fatalf("SyncIgnores: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(repo, "tsconfig.json"))
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("post-patch tsconfig invalid JSON: %v", err)
	}
	var excl []string
	if err := json.Unmarshal(doc["exclude"], &excl); err != nil {
		t.Fatalf("exclude not array: %v", err)
	}
	if len(excl) != 1 || excl[0] != ".worktrees" {
		t.Fatalf("expected exclude=[.worktrees], got %v", excl)
	}
	if _, ok := doc["compilerOptions"]; !ok {
		t.Fatal("compilerOptions key was dropped")
	}
}

func TestSyncIgnores_TsconfigExcludeAsString_RepairsToArray(t *testing.T) {
	repo := t.TempDir()
	cfg := `{"compilerOptions":{},"exclude":"node_modules"}`
	if err := os.WriteFile(filepath.Join(repo, "tsconfig.json"), []byte(cfg), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := worktree.SyncIgnores(repo); err != nil {
		t.Fatalf("SyncIgnores: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(repo, "tsconfig.json"))
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("post-patch tsconfig invalid JSON: %v", err)
	}
	var excl []string
	if err := json.Unmarshal(doc["exclude"], &excl); err != nil {
		t.Fatalf("exclude not coerced to array: %v", err)
	}
	if len(excl) != 1 || excl[0] != ".worktrees" {
		t.Fatalf("expected exclude=[.worktrees] after repair, got %v", excl)
	}
}
