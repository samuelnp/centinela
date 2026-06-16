package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/insights"
	"github.com/samuelnp/centinela/internal/planadvisor"
	"github.com/samuelnp/centinela/internal/telemetry"
)

func seedIntegRepo(t *testing.T, ledger string) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	os.MkdirAll("docs/features", 0o755)                                                                           //nolint:errcheck
	os.MkdirAll(filepath.Join(".workflow", "telemetry"), 0o755)                                                   //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte(fullBrief), 0o644)                                                  //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"f"}]}]}`), 0o644) //nolint:errcheck
	if ledger != "" {
		os.WriteFile(filepath.Join(".workflow", "telemetry", "events.jsonl"), []byte(ledger), 0o644) //nolint:errcheck
	}
}

// fullBrief covers every generic-question topic so the default question cap is
// not exhausted before the gate-failure pre-warning question is reached.
const fullBrief = "## Problem\ntext\n## Scope\nin scope\n## Constraints\nsecurity\n" +
	"## Risks\ntradeoff\n## Acceptance Criteria\nGiven when then\n## Edge Cases\ninvalid input\n"

const integLedger = `{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"import-graph"}
`

func TestIntegPreWarningQuestionNamesWorstGate(t *testing.T) {
	seedIntegRepo(t, integLedger)
	out := planadvisor.Directive("f", &config.Config{})
	if !strings.Contains(out, "recurring gate failures (worst: coverage)") {
		t.Fatalf("expected pre-warning question naming coverage, got:\n%s", out)
	}
	if !strings.Contains(out, "[feature-specialist]") {
		t.Fatalf("expected feature-specialist lens tag on questions, got:\n%s", out)
	}
}

func TestIntegTelemetryDisabledSuppressesFailureContext(t *testing.T) {
	seedIntegRepo(t, integLedger)
	off := false
	cfg := &config.Config{}
	cfg.Telemetry.Enabled = &off
	out := planadvisor.Directive("f", cfg)
	if strings.Contains(out, "recurring gate failures") {
		t.Fatalf("disabled telemetry must suppress all failure context, got:\n%s", out)
	}
}

func TestIntegCountsAgreeWithInsights(t *testing.T) {
	seedIntegRepo(t, integLedger)
	events, err := telemetry.ReadDefault()
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	out := planadvisor.Directive("f", &config.Config{})
	for _, g := range insights.Compute(events, 3).Gates {
		want := fmt.Sprintf("%s (×%d)", g.Key, g.Count)
		if !strings.Contains(out, want) {
			t.Fatalf("advisor output missing %q (must agree with insights), got:\n%s", want, out)
		}
	}
	// Gates and Compute().Gates must be the same ranking (single counter).
	if fmt.Sprintf("%v", insights.Gates(events, 3)) != fmt.Sprintf("%v", insights.Compute(events, 3).Gates) {
		t.Fatal("insights.Gates and Compute().Gates diverge")
	}
}
