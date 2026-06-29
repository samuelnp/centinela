package acceptance_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: Registry resolves a known agent name to its adapter

func TestHostHarnessAC1_RegistryResolvesKnownAgent(t *testing.T) {
	for _, name := range []string{"claude", "opencode", "aider"} {
		a, err := setup.Lookup(name)
		if err != nil {
			t.Fatalf("Lookup(%q): %v", name, err)
		}
		if a.Name() != name {
			t.Fatalf("Lookup(%q).Name() = %q", name, a.Name())
		}
	}
}

// Scenario: Registry returns a typed error for an unknown agent

func TestHostHarnessAC1_RegistryTypedErrorForUnknown(t *testing.T) {
	_, err := setup.Lookup("vscode")
	if !errors.Is(err, setup.ErrUnknownAgent) {
		t.Fatalf("expected ErrUnknownAgent, got %v", err)
	}
	for _, n := range []string{"claude", "opencode", "aider"} {
		if !strings.Contains(err.Error(), n) {
			t.Fatalf("error missing %q: %s", n, err)
		}
	}
}

// Scenario: Claude adapter declares all three capabilities

func TestHostHarnessAC3_ClaudeCapabilities(t *testing.T) {
	a, _ := setup.Lookup("claude")
	caps := a.Capabilities()
	assertHasCap(t, "claude", caps, setup.CapBlocksWrites)
	assertHasCap(t, "claude", caps, setup.CapPromptContext)
	assertHasCap(t, "claude", caps, setup.CapRulesFile)
}

// Scenario: OpenCode adapter declares all three capabilities

func TestHostHarnessAC3_OpenCodeCapabilities(t *testing.T) {
	a, _ := setup.Lookup("opencode")
	caps := a.Capabilities()
	assertHasCap(t, "opencode", caps, setup.CapBlocksWrites)
	assertHasCap(t, "opencode", caps, setup.CapPromptContext)
	assertHasCap(t, "opencode", caps, setup.CapRulesFile)
}

// Scenario: Aider adapter declares prompt-context and rules-file but not blocks-writes

func TestHostHarnessAC3_AiderCapabilities(t *testing.T) {
	a, _ := setup.Lookup("aider")
	caps := a.Capabilities()
	assertHasCap(t, "aider", caps, setup.CapPromptContext)
	assertHasCap(t, "aider", caps, setup.CapRulesFile)
	if hasCap(caps, setup.CapBlocksWrites) {
		t.Fatal("aider must NOT declare blocks-writes")
	}
}

func assertHasCap(t *testing.T, name string, caps []setup.Capability, c setup.Capability) {
	t.Helper()
	if !hasCap(caps, c) {
		t.Fatalf("adapter %q missing capability %q", name, c)
	}
}

func hasCap(caps []setup.Capability, c setup.Capability) bool {
	for _, cap := range caps {
		if cap == c {
			return true
		}
	}
	return false
}
