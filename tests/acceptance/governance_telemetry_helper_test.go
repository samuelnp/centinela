package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// Acceptance: specs/governance-telemetry.feature
//
// Shared helpers for the governance-telemetry acceptance suite. Each test
// chdirs into a temp dir so telemetry writes to an isolated
// .workflow/telemetry/events.jsonl that ReadDefault() reads back.

// gtCfg returns a config with telemetry explicitly enabled or disabled.
func gtCfg(enabled bool) *config.Config {
	c := &config.Config{}
	c.Telemetry = config.TelemetryConfig{Enabled: &enabled}
	return c
}

// gtDefaultCfg returns a zero config — Telemetry.Enabled is nil, which
// IsEnabled() treats as ON (the opt-out default).
func gtDefaultCfg() *config.Config { return &config.Config{} }

// gtChdir isolates telemetry writes to a throwaway working directory.
func gtChdir(t *testing.T) { t.Helper(); t.Chdir(t.TempDir()) }

// gtEvents reads back every recorded event from the current working dir.
func gtEvents(t *testing.T) []telemetry.Event {
	t.Helper()
	evs, err := telemetry.ReadDefault()
	if err != nil {
		t.Fatalf("ReadDefault: %v", err)
	}
	return evs
}
