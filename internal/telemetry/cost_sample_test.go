package telemetry

import "testing"

// RecordCostSample writes a cost-sample line carrying token counts.
func TestRecordCostSampleWrites(t *testing.T) {
	t.Chdir(t.TempDir())
	fixedNow(t, "2026-06-29T12:00:00Z")
	RecordCostSample(enabledCfg(), "cost-governance", "code", "m1", 1500, 1000)

	events, err := ReadDefault()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	e := events[0]
	if e.Type != TypeCostSample || e.InputTokens != 1500 || e.OutputTokens != 1000 || e.Feature != "cost-governance" {
		t.Fatalf("unexpected cost-sample: %+v", e)
	}
}

// A zero/negative total is a no-op so an empty transcript delta writes nothing.
func TestRecordCostSampleZeroIsNoOp(t *testing.T) {
	t.Chdir(t.TempDir())
	RecordCostSample(enabledCfg(), "f", "code", "m", 0, 0)
	events, _ := ReadDefault()
	if len(events) != 0 {
		t.Fatalf("zero total should write nothing, got %d", len(events))
	}
}
