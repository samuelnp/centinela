package setup

import (
	"errors"
	"strings"
	"testing"
)

func TestLookup_KnownAgents(t *testing.T) {
	for _, name := range []string{"claude", "opencode", "aider"} {
		a, err := Lookup(name)
		if err != nil {
			t.Fatalf("Lookup(%q): unexpected error: %v", name, err)
		}
		if a.Name() != name {
			t.Fatalf("Lookup(%q).Name() = %q", name, a.Name())
		}
	}
}

func TestLookup_UnknownAgent_TypedError(t *testing.T) {
	_, err := Lookup("vscode")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	if !errors.Is(err, ErrUnknownAgent) {
		t.Fatalf("expected wrapped ErrUnknownAgent, got %v", err)
	}
	for _, n := range []string{"claude", "opencode", "aider"} {
		if !strings.Contains(err.Error(), n) {
			t.Fatalf("error missing %q: %s", n, err)
		}
	}
}

func TestRegisteredAgents_Order(t *testing.T) {
	agents := RegisteredAgents()
	want := []string{"claude", "opencode", "aider"}
	if len(agents) != len(want) {
		t.Fatalf("want %d agents, got %d", len(want), len(agents))
	}
	for i, w := range want {
		if agents[i] != w {
			t.Fatalf("agents[%d] = %q, want %q", i, agents[i], w)
		}
	}
}

func TestRegisteredAdapters_Count(t *testing.T) {
	if got := RegisteredAdapters(); len(got) != 3 {
		t.Fatalf("expected 3 adapters, got %d", len(got))
	}
}

func TestIsValidAgent(t *testing.T) {
	for _, name := range []string{"claude", "opencode", "aider", "both"} {
		if !IsValidAgent(name) {
			t.Fatalf("IsValidAgent(%q) = false, want true", name)
		}
	}
	if IsValidAgent("vscode") {
		t.Fatal("IsValidAgent(vscode) = true, want false")
	}
}

func TestAgentsFor_Single(t *testing.T) {
	names, err := AgentsFor("aider")
	if err != nil {
		t.Fatalf("AgentsFor(aider): %v", err)
	}
	if len(names) != 1 || names[0] != "aider" {
		t.Fatalf("AgentsFor(aider) = %v, want [aider]", names)
	}
}

func TestAgentsFor_Both(t *testing.T) {
	names, err := AgentsFor("both")
	if err != nil {
		t.Fatalf("AgentsFor(both): %v", err)
	}
	if len(names) != 2 || names[0] != "claude" || names[1] != "opencode" {
		t.Fatalf("AgentsFor(both) = %v, want [claude opencode]", names)
	}
}

func TestAgentsFor_Unknown(t *testing.T) {
	_, err := AgentsFor("unknown")
	if !errors.Is(err, ErrUnknownAgent) {
		t.Fatalf("expected ErrUnknownAgent, got %v", err)
	}
}
