package setup

import (
	"os"
	"testing"
)

// TestInjectHooksNoChangeAndMkdirError covers the !changed early-return (a second
// inject is a no-op) and the MkdirAll error arm (parent path is a regular file).
func TestInjectHooksNoChangeAndMkdirError(t *testing.T) {
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(t.TempDir())

	path := ".claude/settings.json"
	if changed, err := InjectHooks(path); err != nil || !changed {
		t.Fatalf("first inject should create+change: changed=%v err=%v", changed, err)
	}
	if changed, err := InjectHooks(path); err != nil || changed {
		t.Fatalf("second inject should be a no-op: changed=%v err=%v", changed, err)
	}

	os.WriteFile("blocker", []byte("x"), 0644) //nolint:errcheck
	if _, err := InjectHooks("blocker/settings.json"); err == nil {
		t.Fatal("expected MkdirAll error when parent path is a file")
	}
}

// TestInjectOpenCodeConfigNoChange covers the !changed early-return on a second
// merge into an already-complete opencode.json.
func TestInjectOpenCodeConfigNoChange(t *testing.T) {
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(t.TempDir())

	if changed, err := InjectOpenCodeConfig("opencode.json", nil); err != nil || !changed {
		t.Fatalf("first merge should change: changed=%v err=%v", changed, err)
	}
	if changed, err := InjectOpenCodeConfig("opencode.json", nil); err != nil || changed {
		t.Fatalf("second merge should be a no-op: changed=%v err=%v", changed, err)
	}
}

// TestPlanSettingsNoChange covers the !changed (return nil) arm of both
// planHooksSettings and planOpenCodeConfig after the managed files already exist.
func TestPlanSettingsNoChange(t *testing.T) {
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(t.TempDir())

	InjectHooks(".claude/settings.json")  //nolint:errcheck
	InjectOpenCodeConfig("opencode.json", nil) //nolint:errcheck

	if it, err := planHooksSettings(); err != nil || it != nil {
		t.Fatalf("planHooksSettings should be nil no-op: it=%v err=%v", it, err)
	}
	if it, err := planOpenCodeConfig(nil); err != nil || it != nil {
		t.Fatalf("planOpenCodeConfig should be nil no-op: it=%v err=%v", it, err)
	}
}

// TestApplyItemAiderConfigAndAppend covers applyItem's SyncAiderConfig case and
// appendItem's non-nil append branch.
func TestApplyItemAiderConfigAndAppend(t *testing.T) {
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(t.TempDir())

	plan := SyncPlan{Items: []SyncItem{{Kind: SyncAiderConfig, Path: aiderConfigFile, Action: SyncCreate}}}
	if err := ApplySync(plan); err != nil {
		t.Fatalf("apply aider config: %v", err)
	}
	if _, err := os.Stat(aiderConfigFile); err != nil {
		t.Fatalf("expected aider config written: %v", err)
	}

	p := SyncPlan{}
	appendItem(&p, &SyncItem{Kind: SyncAgents, Path: "AGENTS.md", Action: SyncCreate})
	if len(p.Items) != 1 {
		t.Fatalf("expected one appended item, got %d", len(p.Items))
	}
}
