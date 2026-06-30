package main

import (
	"os"
	"testing"
)

// TestRunHarnessSetupClaudeBranch dispatches the "claude" case → setupClaude.
func TestRunHarnessSetupClaudeBranch(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := runHarnessSetup("claude"); err != nil {
		t.Fatalf("claude harness setup: %v", err)
	}
	if _, err := os.Stat(".claude/settings.json"); err != nil {
		t.Fatalf("claude branch should wire settings.json: %v", err)
	}
}

// TestRunHarnessSetupAiderBranch dispatches the "aider" case → setupAider.
func TestRunHarnessSetupAiderBranch(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := runHarnessSetup("aider"); err != nil {
		t.Fatalf("aider harness setup: %v", err)
	}
	if _, err := os.Stat(".aider.conf.yml"); err != nil {
		t.Fatalf("aider branch should write .aider.conf.yml: %v", err)
	}
}

// TestRunHarnessSetupUnknownIsNoOp covers the default case (unknown → nil).
func TestRunHarnessSetupUnknownIsNoOp(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := runHarnessSetup("does-not-exist"); err != nil {
		t.Fatalf("unknown harness should be a no-op, got %v", err)
	}
	if entries, _ := os.ReadDir("."); len(entries) != 0 {
		t.Fatalf("unknown harness must write nothing, got %v", entries)
	}
}

// TestDispatchSetupPropagatesError covers the error branch of dispatchSetup:
// setupOpenCode fails when opencode.json is a directory it cannot read/write.
func TestDispatchSetupPropagatesError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("opencode.json", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := dispatchSetup([]string{"opencode"}); err == nil {
		t.Fatal("dispatchSetup must surface a harness setup error")
	}
}
