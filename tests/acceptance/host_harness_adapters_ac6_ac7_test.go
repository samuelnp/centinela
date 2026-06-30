package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/host-harness-adapters.feature
// Scenario: --agent with a known value is accepted

func TestHostHarnessAC6_KnownAgentIsValid(t *testing.T) {
	for _, agent := range []string{"claude", "opencode", "aider", "both"} {
		if !setup.IsValidAgent(agent) {
			t.Fatalf("IsValidAgent(%q) = false, want true", agent)
		}
	}
}

// Scenario: --agent with an unknown value lists registered harnesses

func TestHostHarnessAC6_UnknownAgentExitsWithHarnessList(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	out, code := runCent(t, bin, dir, "init", "--agent", "unknown-tool")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown --agent\noutput:\n%s", out)
	}
	for _, name := range setup.RegisteredAgents() {
		if !strings.Contains(out, name) {
			t.Fatalf("error output missing harness %q:\n%s", name, out)
		}
	}
}

// Scenario: isValidAgent is resolved by the registry, not a hardcoded list

func TestHostHarnessAC6_IsValidAgent_RegistryDriven(t *testing.T) {
	for _, name := range setup.RegisteredAgents() {
		if !setup.IsValidAgent(name) {
			t.Fatalf("IsValidAgent(%q) = false, want true", name)
		}
	}
	if setup.IsValidAgent("cursor") {
		t.Fatal("IsValidAgent(cursor) = true, want false")
	}
}

// Scenario: Every registered adapter declares a non-empty capability set

func TestHostHarnessAC7_AllAdaptersNonEmptyCapabilities(t *testing.T) {
	for _, a := range setup.RegisteredAdapters() {
		if len(a.Capabilities()) == 0 {
			t.Fatalf("adapter %q has empty capabilities", a.Name())
		}
	}
}
