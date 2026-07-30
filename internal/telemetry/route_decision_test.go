package telemetry

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestRecordRouteDecision_WritesTheAuditLine(t *testing.T) {
	dir := t.TempDir()
	origin, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origin) }) //nolint:errcheck
	os.Chdir(dir)                          //nolint:errcheck

	cfg := &config.Config{}
	cfg.Telemetry.Enabled = boolPtr(true)
	RecordRouteDecision(cfg, "f", "senior-engineer", "balanced", "reasoning", "config-only change", "opus")

	data, err := os.ReadFile(eventsFile)
	if err != nil {
		t.Fatalf("expected an events file: %v", err)
	}
	for _, want := range []string{
		`"type":"route-decision"`, `"role":"senior-engineer"`, `"tier":"balanced"`,
		`"prevTier":"reasoning"`, `"reason":"config-only change"`, `"model":"opus"`,
	} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("event missing %s: %s", want, data)
		}
	}
}

func boolPtr(b bool) *bool { return &b }
