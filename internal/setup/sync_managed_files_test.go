package setup

import (
	"os"
	"testing"
)

func TestPlanManagedFileBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	item, err := planManagedFile("x", "target", "legacy", SyncAgents)
	if err != nil || item.Action != SyncCreate {
		t.Fatalf("expected create, got %v %v", item, err)
	}
	os.WriteFile("x", []byte("target"), 0644) //nolint:errcheck
	item, err = planManagedFile("x", "target", "legacy", SyncAgents)
	if err != nil || item != nil {
		t.Fatalf("expected nil for already target, got %v %v", item, err)
	}
	os.WriteFile("x", []byte("legacy"), 0644) //nolint:errcheck
	item, _ = planManagedFile("x", "target", "legacy", SyncAgents)
	if item.Action != SyncUpdate {
		t.Fatal("expected update for legacy content")
	}
	os.WriteFile("x", []byte("<!-- centinela:managed-version=0 -->\ncustom"), 0644) //nolint:errcheck
	item, _ = planManagedFile("x", "target", "legacy", SyncAgents)
	if item.Action != SyncUpdate {
		t.Fatal("expected update for managed header content")
	}
	os.WriteFile("x", []byte("custom"), 0644) //nolint:errcheck
	item, _ = planManagedFile("x", "target", "legacy", SyncAgents)
	if item.Action != SyncManualReview {
		t.Fatal("expected manual-review for unmanaged content")
	}
}
