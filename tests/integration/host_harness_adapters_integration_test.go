package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: centinela init --agent aider writes Aider managed files

func TestHostHarnessAider_InitWritesManagedFiles(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, err := setup.BuildSyncPlan("aider")
	if err != nil {
		t.Fatalf("BuildSyncPlan(aider): %v", err)
	}
	if err := setup.ApplySync(plan); err != nil {
		t.Fatalf("ApplySync: %v", err)
	}
	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("AGENTS.md missing: %v", err)
	}
	if !strings.Contains(string(data), "centinela:managed-version=") {
		t.Fatal("AGENTS.md missing managed header")
	}
	cfg, err := os.ReadFile(".aider.conf.yml")
	if err != nil {
		t.Fatalf(".aider.conf.yml missing: %v", err)
	}
	if !strings.Contains(string(cfg), "read: AGENTS.md") {
		t.Fatalf(".aider.conf.yml missing read: AGENTS.md:\n%s", cfg)
	}
}

// Scenario: centinela init --agent aider is idempotent on re-run

func TestHostHarnessAider_IdempotentReApply(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	plan, _ := setup.BuildSyncPlan("aider")
	setup.ApplySync(plan) //nolint:errcheck

	cfg1, _ := os.ReadFile(".aider.conf.yml")
	md1, _ := os.ReadFile("AGENTS.md")

	plan2, _ := setup.BuildSyncPlan("aider")
	setup.ApplySync(plan2) //nolint:errcheck

	cfg2, _ := os.ReadFile(".aider.conf.yml")
	md2, _ := os.ReadFile("AGENTS.md")

	if string(cfg1) != string(cfg2) {
		t.Fatal(".aider.conf.yml changed on second apply")
	}
	if string(md1) != string(md2) {
		t.Fatal("AGENTS.md changed on second apply")
	}
}

// Scenario: Partial existing install - adding Aider leaves Claude files untouched

func TestHostHarnessAider_DoesNotTouchClaudeFiles(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	os.MkdirAll(".claude", 0755) //nolint:errcheck
	const claudeContent = `{"hooks": {}}`
	os.WriteFile(".claude/settings.json", []byte(claudeContent), 0644) //nolint:errcheck

	plan, _ := setup.BuildSyncPlan("aider")
	setup.ApplySync(plan) //nolint:errcheck

	got, _ := os.ReadFile(".claude/settings.json")
	if string(got) != claudeContent {
		t.Fatal(".claude/settings.json was modified by aider init")
	}
}
