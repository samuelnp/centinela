package setup

import (
	"testing"
)

func hasAdapterCapability(caps []Capability, c Capability) bool {
	for _, cap := range caps {
		if cap == c {
			return true
		}
	}
	return false
}

func TestClaudeAdapter_NameAndCapabilities(t *testing.T) {
	a := claudeAdapter{}
	if a.Name() != "claude" {
		t.Fatalf("Name() = %q, want claude", a.Name())
	}
	caps := a.Capabilities()
	for _, c := range []Capability{CapBlocksWrites, CapPromptContext, CapRulesFile} {
		if !hasAdapterCapability(caps, c) {
			t.Fatalf("claude missing capability %q", c)
		}
	}
}

func TestOpenCodeAdapter_NameAndCapabilities(t *testing.T) {
	a := openCodeAdapter{}
	if a.Name() != "opencode" {
		t.Fatalf("Name() = %q, want opencode", a.Name())
	}
	caps := a.Capabilities()
	for _, c := range []Capability{CapBlocksWrites, CapPromptContext, CapRulesFile} {
		if !hasAdapterCapability(caps, c) {
			t.Fatalf("opencode missing capability %q", c)
		}
	}
}

func TestAiderAdapter_NameAndCapabilities(t *testing.T) {
	a := aiderAdapter{}
	if a.Name() != "aider" {
		t.Fatalf("Name() = %q, want aider", a.Name())
	}
	caps := a.Capabilities()
	if !hasAdapterCapability(caps, CapPromptContext) {
		t.Fatal("aider missing prompt-context")
	}
	if !hasAdapterCapability(caps, CapRulesFile) {
		t.Fatal("aider missing rules-file")
	}
	if hasAdapterCapability(caps, CapBlocksWrites) {
		t.Fatal("aider must NOT declare blocks-writes")
	}
}
