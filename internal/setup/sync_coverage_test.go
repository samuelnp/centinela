package setup

import (
	"os"
	"testing"
)

func TestBuildSyncPlanErrorAndApplyItemBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("AGENTS.md", 0755) //nolint:errcheck
	if _, err := BuildSyncPlan("opencode"); err == nil {
		t.Fatal("expected sync plan error when AGENTS.md is a directory")
	}

	os.RemoveAll("AGENTS.md") //nolint:errcheck
	plan := SyncPlan{Items: []SyncItem{
		{Kind: SyncClaudeHooks, Path: ".claude/settings.json", Action: SyncCreate},
		{Kind: SyncOpenCodeCfg, Path: "opencode.json", Action: SyncCreate},
	}}
	if err := ApplySync(plan); err != nil {
		t.Fatalf("expected apply item branches to pass: %v", err)
	}

	p := SyncPlan{}
	appendItem(&p, nil)
	if p.HasChanges() {
		t.Fatal("expected nil append to keep empty plan")
	}
}
