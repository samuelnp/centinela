package acceptance_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: BuildSyncPlan for "claude" produces only Claude managed items

func TestHostHarnessAC2_ClaudeScope(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, err := setup.BuildSyncPlan("claude")
	if err != nil {
		t.Fatalf("BuildSyncPlan(claude): %v", err)
	}
	hasSettings := false
	for _, it := range plan.Items {
		if it.Path == ".claude/settings.json" {
			hasSettings = true
		}
		if it.Path == "opencode.json" || it.Path == "AGENTS.md" || it.Path == ".aider.conf.yml" {
			t.Fatalf("claude plan has unexpected item: %s", it.Path)
		}
	}
	if !hasSettings {
		t.Fatal("claude plan missing .claude/settings.json")
	}
}

// Scenario: BuildSyncPlan for "aider" produces only Aider managed items

func TestHostHarnessAC2_AiderScope(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, err := setup.BuildSyncPlan("aider")
	if err != nil {
		t.Fatalf("BuildSyncPlan(aider): %v", err)
	}
	hasAiderCfg, hasAgentsMd := false, false
	for _, it := range plan.Items {
		switch it.Path {
		case ".aider.conf.yml":
			hasAiderCfg = true
		case "AGENTS.md":
			hasAgentsMd = true
		case ".claude/settings.json", "opencode.json":
			t.Fatalf("aider plan has cross-harness item: %s", it.Path)
		}
	}
	if !hasAiderCfg {
		t.Fatal("aider plan missing .aider.conf.yml")
	}
	if !hasAgentsMd {
		t.Fatal("aider plan missing AGENTS.md")
	}
}

// Scenario: BuildSyncPlan for "both" composes Claude and OpenCode items

func TestHostHarnessAC2_BothComposesClaudeOpenCode(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	both, err := setup.BuildSyncPlan("both")
	if err != nil {
		t.Fatalf("BuildSyncPlan(both): %v", err)
	}
	claude, _ := setup.BuildSyncPlan("claude")
	opencode, _ := setup.BuildSyncPlan("opencode")
	want := len(claude.Items) + len(opencode.Items)
	if len(both.Items) != want {
		t.Fatalf("both: got %d items, want %d (claude=%d opencode=%d)",
			len(both.Items), want, len(claude.Items), len(opencode.Items))
	}
	for _, it := range both.Items {
		if it.Path == ".aider.conf.yml" {
			t.Fatal("both plan must not contain aider config")
		}
	}
}
