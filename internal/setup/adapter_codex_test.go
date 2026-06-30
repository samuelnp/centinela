package setup

import (
	"os"
	"testing"
)

func TestCodexAdapter_NameAndCapabilities(t *testing.T) {
	a := codexAdapter{}
	if a.Name() != "codex" {
		t.Fatalf("Name() = %q, want codex", a.Name())
	}
	caps := a.Capabilities()
	for _, c := range []Capability{CapBlocksWrites, CapPromptContext, CapRulesFile} {
		if !hasAdapterCapability(caps, c) {
			t.Fatalf("codex missing capability %q", c)
		}
	}
}

func TestCodexAdapter_PlanItems(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	items, err := codexAdapter{}.PlanItems()
	if err != nil {
		t.Fatalf("PlanItems: %v", err)
	}
	var cfg, agents bool
	for _, it := range items {
		if it.Path == codexConfigFile {
			cfg = true
			if it.Kind != SyncKindPrewriteHook {
				t.Fatalf("config kind = %s, want %s", it.Kind, SyncKindPrewriteHook)
			}
		}
		if it.Path == "AGENTS.md" && it.Kind == SyncAgents {
			agents = true
		}
	}
	if !cfg {
		t.Fatalf("PlanItems missing %s prewrite-hook item: %+v", codexConfigFile, items)
	}
	if !agents {
		t.Fatalf("PlanItems missing AGENTS.md (SyncAgents) item: %+v", items)
	}
}
