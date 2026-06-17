package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestModelFieldSerializesWhenSet — a stamped Model serializes into the JSON line
// under the "model" key.
func TestModelFieldSerializesWhenSet(t *testing.T) {
	line, err := json.Marshal(Event{Type: TypeStepAdvanced, Model: "claude-sonnet-4-6"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(line), `"model":"claude-sonnet-4-6"`) {
		t.Fatalf("model not serialized: %s", line)
	}
}

// TestModelFieldOmittedWhenEmpty — an empty Model is omitted (omitempty) so legacy
// readers and golden lines stay stable.
func TestModelFieldOmittedWhenEmpty(t *testing.T) {
	line, err := json.Marshal(Event{Type: TypeStepAdvanced})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(line), "model") {
		t.Fatalf("empty model should be omitted: %s", line)
	}
}

// TestLegacyLineWithoutModelParsesEmpty — a back-compat line with no "model" key
// unmarshals cleanly with Model == "".
func TestLegacyLineWithoutModelParsesEmpty(t *testing.T) {
	const legacy = `{"schema":"centinela.telemetry/v1","type":"step-advanced","timestamp":"2026-01-01T00:00:00Z","feature":"alpha","step":"plan"}`
	var e Event
	if err := json.Unmarshal([]byte(legacy), &e); err != nil {
		t.Fatalf("legacy line should parse: %v", err)
	}
	if e.Model != "" {
		t.Fatalf("legacy Model should be empty, got %q", e.Model)
	}
	if e.Type != TypeStepAdvanced || e.Feature != "alpha" {
		t.Fatalf("legacy fields lost: %+v", e)
	}
}

// TestConstructorsStampModel — every Record* constructor stamps the passed model
// onto the appended event.
func TestConstructorsStampModel(t *testing.T) {
	t.Chdir(t.TempDir())
	const m = "claude-opus-4-7"
	cfg := enabledCfg()
	RecordBlock(cfg, "f", "plan", "plan", "p", "out-of-step", m)
	RecordGateFailure(cfg, "G1", "msg", m)
	RecordVerifyRejection(cfg, "f", "validate", []CheckRef{{Claim: "c"}}, m)
	RecordCompleteRejected(cfg, "f", "validate", "gates", m)
	RecordStepAdvanced(cfg, "f", "plan", m)
	evs, err := Read(telemetryDir)
	if err != nil || len(evs) != 5 {
		t.Fatalf("read: %v len=%d", err, len(evs))
	}
	for _, e := range evs {
		if e.Model != m {
			t.Fatalf("%s did not stamp model: %+v", e.Type, e)
		}
	}
}
