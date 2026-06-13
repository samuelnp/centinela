package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/model-capability-profiles.feature

func mcpLoad(t *testing.T, toml string) (*config.Config, error) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.WriteFile(config.Filename, []byte(toml), 0644); err != nil {
		t.Fatalf("write toml: %v", err)
	}
	return config.Load()
}

func mcpLoadFailsNaming(t *testing.T, toml, substr string) {
	t.Helper()
	if _, err := mcpLoad(t, toml); err == nil || !strings.Contains(err.Error(), substr) {
		t.Fatalf("want load error naming %q, got %v", substr, err)
	}
}

// Scenario: Capability class values are normalized by trim and lowercase
func TestMCP_ClassValuesNormalized(t *testing.T) {
	cfg, err := mcpLoad(t, "[orchestration.capabilities]\n\"local/m\" = \"  Frontier  \"\n")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if class, ok := config.CapabilityClassFor("local/m", cfg); !ok || class != config.CapabilityFrontier {
		t.Fatalf("local/m = (%q,%v), want (frontier,true)", class, ok)
	}
}

// Scenario: An unknown capability class value fails config load
func TestMCP_UnknownClassValueFailsLoad(t *testing.T) {
	mcpLoadFailsNaming(t, "[orchestration.capabilities]\n\"local/m\" = \"genius\"\n", "genius")
}

// Scenario: An empty model id key in capabilities fails config load
func TestMCP_EmptyModelIDFailsLoad(t *testing.T) {
	mcpLoadFailsNaming(t, "[orchestration.capabilities]\n\"\" = \"frontier\"\n", "empty")
}

// Scenario: An unknown class key in capability_profiles fails config load
func TestMCP_UnknownClassKeyFailsLoad(t *testing.T) {
	mcpLoadFailsNaming(t, "[orchestration.capability_profiles]\ngenius = \"guided\"\n", "genius")
}

// Scenario: An unknown profile value in capability_profiles fails config load
func TestMCP_UnknownProfileValueFailsLoad(t *testing.T) {
	mcpLoadFailsNaming(t, "[orchestration.capability_profiles]\nfrontier = \"turbo\"\n", "turbo")
}

// Scenario: Absent capability tables are valid and change nothing
func TestMCP_AbsentTablesValid(t *testing.T) {
	cfg, err := mcpLoad(t, "[workflow]\n")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := workflow.EffectiveProfile(&workflow.Workflow{}, cfg); got != config.ProfileStrict {
		t.Fatalf("zero-config effective = %q, want strict", got)
	}
}
