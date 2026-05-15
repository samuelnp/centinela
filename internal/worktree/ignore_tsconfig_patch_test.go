package worktree_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

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
