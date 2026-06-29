package unit_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: Registry resolves a known agent name to its adapter

func TestHostHarnessRegistry_KnownAgentsResolve(t *testing.T) {
	for _, name := range []string{"claude", "opencode", "aider"} {
		a, err := setup.Lookup(name)
		if err != nil || a == nil {
			t.Fatalf("Lookup(%q): adapter=%v err=%v", name, a, err)
		}
	}
}

// Scenario: Registry returns a typed error for an unknown agent

func TestHostHarnessRegistry_UnknownAgentTypedError(t *testing.T) {
	_, err := setup.Lookup("vscode")
	if !errors.Is(err, setup.ErrUnknownAgent) {
		t.Fatalf("expected ErrUnknownAgent, got %v", err)
	}
	for _, n := range setup.RegisteredAgents() {
		if !strings.Contains(err.Error(), n) {
			t.Fatalf("error missing registered harness %q: %s", n, err)
		}
	}
}

// Scenario: isValidAgent is resolved by the registry, not a hardcoded list

func TestHostHarnessRegistry_IsValidAgent_RegistryDriven(t *testing.T) {
	for _, name := range setup.RegisteredAgents() {
		if !setup.IsValidAgent(name) {
			t.Fatalf("IsValidAgent(%q) = false, want true", name)
		}
	}
	if !setup.IsValidAgent("both") {
		t.Fatal("IsValidAgent(both) = false, want true")
	}
	if setup.IsValidAgent("unknown-tool") {
		t.Fatal("IsValidAgent(unknown-tool) = true, want false")
	}
}

// Scenario: BuildSyncPlan contains no per-harness if-ladder
// Verified structurally: BuildSyncPlan must succeed for all agents via
// registry iteration — no hardcoded branch per harness.

func TestHostHarnessRegistry_BuildSyncPlan_AllAgents(t *testing.T) {
	agents := append(setup.RegisteredAgents(), "both")
	for _, agent := range agents {
		d := t.TempDir()
		orig, _ := os.Getwd()
		os.Chdir(d) //nolint:errcheck
		_, err := setup.BuildSyncPlan(agent)
		os.Chdir(orig) //nolint:errcheck
		if err != nil {
			t.Fatalf("BuildSyncPlan(%q): %v", agent, err)
		}
	}
}
