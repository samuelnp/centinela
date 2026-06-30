package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: centinela init --agent aider writes Aider managed files

func TestHostHarnessAC5_AiderInitWritesFiles(t *testing.T) {
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
	cfg, err := os.ReadFile(".aider.conf.yml")
	if err != nil {
		t.Fatalf(".aider.conf.yml not created: %v", err)
	}
	if !strings.Contains(string(cfg), "read: AGENTS.md") {
		t.Fatalf(".aider.conf.yml missing read: AGENTS.md:\n%s", cfg)
	}
	if _, err := os.Stat(".claude/settings.json"); err == nil {
		t.Fatal(".claude/settings.json must not be created by aider init")
	}
	if _, err := os.Stat("opencode.json"); err == nil {
		t.Fatal("opencode.json must not be created by aider init")
	}
}

// Scenario: centinela init --agent aider is idempotent on re-run

func TestHostHarnessAC5_AiderInitIdempotent(t *testing.T) {
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

// Scenario: Pre-existing unmanaged .aider.conf.yml is not clobbered

func TestHostHarnessAC5_UnmanagedAiderConfigNotClobbered(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	const userCfg = "# user config\nread: myfile.md\n"
	os.WriteFile(".aider.conf.yml", []byte(userCfg), 0644) //nolint:errcheck

	plan, err := setup.BuildSyncPlan("aider")
	if err != nil {
		t.Fatalf("BuildSyncPlan: %v", err)
	}
	var hasManualReview bool
	for _, it := range plan.Items {
		if it.Path == ".aider.conf.yml" && it.Action == setup.SyncManualReview {
			hasManualReview = true
		}
	}
	if !hasManualReview {
		t.Fatal("expected manual-review item for unmanaged .aider.conf.yml")
	}
	setup.ApplySync(plan) //nolint:errcheck
	got, _ := os.ReadFile(".aider.conf.yml")
	if string(got) != userCfg {
		t.Fatal("unmanaged .aider.conf.yml was overwritten")
	}
}
