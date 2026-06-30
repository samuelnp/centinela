package unit_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Scenario: BuildSyncPlan for "claude" produces only Claude managed items

func TestHostHarnessScopeClaude_OnlyClaudeItems(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, err := setup.BuildSyncPlan("claude")
	if err != nil {
		t.Fatalf("BuildSyncPlan(claude): %v", err)
	}
	for _, it := range plan.Items {
		if it.Path == "opencode.json" || it.Path == "AGENTS.md" || it.Path == ".aider.conf.yml" {
			t.Fatalf("claude plan has cross-harness item: %s", it.Path)
		}
	}
}

// Scenario: BuildSyncPlan for "opencode" produces only OpenCode managed items

func TestHostHarnessScopeOpenCode_OnlyOpenCodeItems(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, err := setup.BuildSyncPlan("opencode")
	if err != nil {
		t.Fatalf("BuildSyncPlan(opencode): %v", err)
	}
	for _, it := range plan.Items {
		if it.Path == ".claude/settings.json" || it.Path == ".aider.conf.yml" {
			t.Fatalf("opencode plan has cross-harness item: %s", it.Path)
		}
	}
}

// Scenario: BuildSyncPlan for "aider" produces only Aider managed items

func TestHostHarnessScopeAider_OnlyAiderItems(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, err := setup.BuildSyncPlan("aider")
	if err != nil {
		t.Fatalf("BuildSyncPlan(aider): %v", err)
	}
	for _, it := range plan.Items {
		if it.Path == ".claude/settings.json" || it.Path == "opencode.json" {
			t.Fatalf("aider plan has cross-harness item: %s", it.Path)
		}
	}
}

// Scenario: BuildSyncPlan for "both" composes Claude and OpenCode items

func TestHostHarnessScopeBoth_UnionOfClaudeAndOpenCode(t *testing.T) {
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
		t.Fatalf("both has %d items, want %d (claude=%d + opencode=%d)",
			len(both.Items), want, len(claude.Items), len(opencode.Items))
	}
	for _, it := range both.Items {
		if it.Path == ".aider.conf.yml" {
			t.Fatal("both plan must not contain aider items")
		}
	}
}
