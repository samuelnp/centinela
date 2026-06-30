package setup

import (
	"os"
	"testing"
)

func TestCapabilityParity_AllNonEmpty(t *testing.T) {
	for _, a := range RegisteredAdapters() {
		if len(a.Capabilities()) == 0 {
			t.Fatalf("adapter %q has empty capabilities", a.Name())
		}
	}
}

func TestCapabilityParity_BlocksWritesRequiresPrewriteHook(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	for _, a := range RegisteredAdapters() {
		if !hasAdapterCapability(a.Capabilities(), CapBlocksWrites) {
			continue
		}
		items, err := a.PlanItems()
		if err != nil {
			t.Fatalf("adapter %q PlanItems(): %v", a.Name(), err)
		}
		found := false
		for _, it := range items {
			if it.Kind == SyncKindPrewriteHook {
				found = true
			}
		}
		if !found {
			t.Fatalf("adapter %q claims blocks-writes but no prewrite hook item", a.Name())
		}
	}
}

func TestAiderAdapter_NoPrewriteHook(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	a := aiderAdapter{}
	items, err := a.PlanItems()
	if err != nil {
		t.Fatalf("PlanItems(): %v", err)
	}
	for _, it := range items {
		if it.Kind == SyncKindPrewriteHook {
			t.Fatal("aider must not produce a prewrite hook item")
		}
	}
}
