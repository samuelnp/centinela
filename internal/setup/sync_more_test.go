package setup

import (
	"os"
	"testing"
)

func TestSyncPlanHasChangesAndClassifyAction(t *testing.T) {
	if (SyncPlan{}).HasChanges() {
		t.Fatal("expected empty plan no changes")
	}
	if classifyAction("missing") != SyncCreate {
		t.Fatal("expected create action for missing path")
	}
	d := t.TempDir()
	path := d + "/x"
	os.WriteFile(path, []byte("x"), 0644) //nolint:errcheck
	if classifyAction(path) != SyncUpdate {
		t.Fatal("expected update action for existing path")
	}
}

func TestBuildSyncPlanByAgent(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	claude, _ := BuildSyncPlan("claude")
	for _, it := range claude.Items {
		if it.Kind != SyncKindPrewriteHook {
			t.Fatalf("unexpected kind for claude scope: %s", it.Kind)
		}
	}
	opencode, _ := BuildSyncPlan("opencode")
	for _, it := range opencode.Items {
		if it.Path == ".claude/settings.json" {
			t.Fatal("unexpected claude settings in opencode scope")
		}
	}
}

func TestApplySyncManualReviewUnknownAndError(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("AGENTS.md", []byte("custom"), 0644) //nolint:errcheck
	plan := SyncPlan{Items: []SyncItem{{Kind: SyncAgents, Path: "AGENTS.md", Action: SyncManualReview}}}
	if err := ApplySync(plan); err != nil {
		t.Fatal(err)
	}
	if err := ApplySync(SyncPlan{Items: []SyncItem{{Kind: "unknown", Path: "x", Action: SyncCreate}}}); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(".opencode", []byte("x"), 0644) //nolint:errcheck
	plug := SyncPlan{Items: []SyncItem{{Kind: SyncKindPrewriteHook, Path: pluginFile, Action: SyncUpdate}}}
	if err := ApplySync(plug); err == nil {
		t.Fatal("expected plugin write error with conflicting .opencode file")
	}
}
