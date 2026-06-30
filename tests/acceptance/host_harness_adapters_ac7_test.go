package acceptance_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: Any adapter claiming blocks-writes wires a prewrite hook

func TestHostHarnessAC7_BlocksWritesRequiresPrewriteHook(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	for _, a := range setup.RegisteredAdapters() {
		if !hasCap(a.Capabilities(), setup.CapBlocksWrites) {
			continue
		}
		items, err := a.PlanItems()
		if err != nil {
			t.Fatalf("adapter %q PlanItems(): %v", a.Name(), err)
		}
		found := false
		for _, it := range items {
			if it.Kind == setup.SyncKindPrewriteHook {
				found = true
			}
		}
		if !found {
			t.Fatalf("adapter %q claims blocks-writes but no prewrite hook item", a.Name())
		}
	}
}

// Scenario: Aider does not wire a prewrite hook

func TestHostHarnessAC7_AiderNoPrewriteHook(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	a, _ := setup.Lookup("aider")
	items, err := a.PlanItems()
	if err != nil {
		t.Fatalf("PlanItems(): %v", err)
	}
	for _, it := range items {
		if it.Kind == setup.SyncKindPrewriteHook {
			t.Fatal("aider must not produce a prewrite hook item")
		}
	}
}

// Scenario: Hook-less harness cannot claim blocks-writes
// This assertion is enforced by the parity test above: any adapter with
// blocks-writes that lacks a SyncKindPrewriteHook item fails the test.
// The invariant is a compile-time contract between adapters and the registry.

func TestHostHarnessAC7_HooklessHarnessCannotClaimBlocksWrites(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	aider, _ := setup.Lookup("aider")
	if hasCap(aider.Capabilities(), setup.CapBlocksWrites) {
		t.Fatal("aider declares blocks-writes but has no prewrite hook — violates parity")
	}
	items, err := aider.PlanItems()
	if err != nil {
		t.Fatalf("aider PlanItems(): %v", err)
	}
	for _, it := range items {
		if it.Kind == setup.SyncKindPrewriteHook {
			t.Fatal("aider must not emit a prewrite hook item")
		}
	}
}
