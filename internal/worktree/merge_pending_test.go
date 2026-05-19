package worktree_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestPendingPath_NamesMarkerUnderWorkflow(t *testing.T) {
	p := worktree.PendingPath("/repo", "zeta")
	if !strings.HasSuffix(p, filepath.Join(".workflow", "zeta-merge-pending.json")) {
		t.Fatalf("PendingPath = %q", p)
	}
}

func TestWritePending_AtomicNoLeftoverTmp(t *testing.T) {
	d := t.TempDir()
	o := worktree.MergeOutcome{Feature: "delta", TextConflict: true,
		ConflictedPaths: []string{"a.txt"}}
	if err := worktree.WritePending(d, o); err != nil {
		t.Fatalf("WritePending: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(d, "delta") + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file must not survive WritePending; stat err=%v", err)
	}
	data, err := os.ReadFile(worktree.PendingPath(d, "delta"))
	if err != nil {
		t.Fatalf("marker missing: %v", err)
	}
	var m worktree.PendingMarker
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("marker is not well-formed JSON: %v", err)
	}
	if m.Reason != "git-text-conflict" || m.Feature != "delta" {
		t.Fatalf("unexpected marker: %+v", m)
	}
	if !strings.HasSuffix(m.WorktreePath, filepath.Join(".worktrees", "delta")) {
		t.Fatalf("worktree path not recorded: %q", m.WorktreePath)
	}
}

func TestWritePending_IdempotentRewriteReasonReplacedNotAppended(t *testing.T) {
	d := t.TempDir()
	first := worktree.MergeOutcome{Feature: "omicron", TextConflict: true}
	if err := worktree.WritePending(d, first); err != nil {
		t.Fatalf("first WritePending: %v", err)
	}
	second := worktree.MergeOutcome{Feature: "omicron", ValidateFail: true}
	if err := worktree.WritePending(d, second); err != nil {
		t.Fatalf("second WritePending: %v", err)
	}
	data, _ := os.ReadFile(worktree.PendingPath(d, "omicron"))
	// Exactly one JSON object — a re-stall rewrites, never appends.
	if strings.Count(string(data), "\"feature\"") != 1 {
		t.Fatalf("marker appended instead of rewritten: %s", data)
	}
	m, err := worktree.LoadPending(d, "omicron")
	if err != nil || m == nil {
		t.Fatalf("LoadPending: %v / %v", m, err)
	}
	if m.Reason != "post-merge-validate-failed" {
		t.Fatalf("reason not replaced on re-stall: %q", m.Reason)
	}
}

func TestLoadPending_MissingReturnsNilNil(t *testing.T) {
	m, err := worktree.LoadPending(t.TempDir(), "theta")
	if m != nil || err != nil {
		t.Fatalf("absent marker must be (nil,nil), got (%v,%v)", m, err)
	}
}

func TestLoadPending_CorruptJSONReturnsError(t *testing.T) {
	d := t.TempDir()
	_ = os.MkdirAll(filepath.Join(d, ".workflow"), 0o755)
	_ = os.WriteFile(worktree.PendingPath(d, "bad"), []byte("{not json"), 0o644)
	m, err := worktree.LoadPending(d, "bad")
	if err == nil {
		t.Fatal("corrupt marker must return an error, not nil")
	}
	if m != nil {
		t.Fatalf("corrupt marker must not yield a marker, got %+v", m)
	}
}
