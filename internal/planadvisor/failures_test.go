package planadvisor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func seedLedger(t *testing.T, lines string) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if lines == "" {
		return
	}
	dir := filepath.Join(".workflow", "telemetry")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func enabledCfg() *config.Config {
	on := true
	c := &config.Config{}
	c.Telemetry.Enabled = &on
	return c
}

func disabledCfg() *config.Config {
	off := false
	c := &config.Config{}
	c.Telemetry.Enabled = &off
	return c
}

func TestRecurringFailuresRanksSeededLedger(t *testing.T) {
	seedLedger(t, `{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"import-graph"}
`)
	got := recurringFailures(enabledCfg(), 5)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].Key != "coverage" || got[0].Count != 2 {
		t.Fatalf("first = %+v", got[0])
	}
	if got[1].Key != "import-graph" || got[1].Count != 1 {
		t.Fatalf("second = %+v", got[1])
	}
}

func TestRecurringFailuresTelemetryDisabledNil(t *testing.T) {
	seedLedger(t, `{"type":"gate-failure","gate":"coverage"}`)
	if got := recurringFailures(disabledCfg(), 5); got != nil {
		t.Fatalf("disabled telemetry must yield nil, got %+v", got)
	}
}

func TestRecurringFailuresMissingLedgerEmpty(t *testing.T) {
	seedLedger(t, "")
	if got := recurringFailures(enabledCfg(), 5); len(got) != 0 {
		t.Fatalf("missing ledger must yield no failures, got %+v", got)
	}
}

func TestRecurringFailuresTopNZeroNil(t *testing.T) {
	seedLedger(t, `{"type":"gate-failure","gate":"coverage"}`)
	if got := recurringFailures(enabledCfg(), 0); got != nil {
		t.Fatalf("topN<=0 must yield nil, got %+v", got)
	}
	if got := recurringFailures(nil, 5); got != nil {
		t.Fatalf("nil cfg must yield nil, got %+v", got)
	}
}

func TestFailureTopNDefaultsAndClamps(t *testing.T) {
	if failureTopN(nil) != 3 {
		t.Fatalf("nil cfg topN = %d, want 3", failureTopN(nil))
	}
	c := &config.Config{}
	c.Workflow.PlanAdvisorFailureTopN = 9
	if got := failureTopN(c); got != 5 {
		t.Fatalf("clamp high topN = %d, want 5", got)
	}
	c.Workflow.PlanAdvisorFailureTopN = -2
	if got := failureTopN(c); got != 3 {
		t.Fatalf("clamp low topN = %d, want 3", got)
	}
}
