package acceptance_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: BuildSyncPlan for "opencode" produces only OpenCode managed items

func TestHostHarnessAC2_OpenCodeScope(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, err := setup.BuildSyncPlan("opencode")
	if err != nil {
		t.Fatalf("BuildSyncPlan(opencode): %v", err)
	}
	hasConfig := false
	for _, it := range plan.Items {
		if it.Path == "opencode.json" {
			hasConfig = true
		}
		if it.Path == ".claude/settings.json" || it.Path == ".aider.conf.yml" {
			t.Fatalf("opencode plan has cross-harness item: %s", it.Path)
		}
	}
	if !hasConfig {
		t.Fatal("opencode plan missing opencode.json")
	}
}
