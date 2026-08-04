package workflow

import (
	"os"
	"testing"
)

// futureFiles are state files a NEWER Centinela could plausibly write: ones
// that changed the TYPE of schemaVersion itself, and ones that changed the
// SHAPE of a field this binary models. None of them can be unmarshalled into
// Workflow, and before the version probe every one of them made Load fail —
// which empties ActiveWorkflows, turns EvaluatePrewrite into NeedInit, blocks
// every governed write and makes hook_autostart fork a duplicate workflow.
var futureFiles = map[string]string{
	"version is a string":   `{"schemaVersion":"2.0","feature":"delta","currentStep":"code","steps":{}}`,
	"version is a float":    `{"schemaVersion":2.0,"feature":"delta","currentStep":"code","steps":{}}`,
	"version is fractional": `{"schemaVersion":1.5,"feature":"delta","currentStep":"code","steps":{}}`,
	"version is exponent":   `{"schemaVersion":2e2,"feature":"delta","currentStep":"code","steps":{}}`,
	"version is a bool":     `{"schemaVersion":true,"feature":"delta","currentStep":"code","steps":{}}`,
	"version overflows int": `{"schemaVersion":99999999999999999999,"feature":"delta","currentStep":"code","steps":{}}`,
	"v99 reshapes steps":    `{"schemaVersion":99,"feature":"delta","currentStep":"code","steps":[{"name":"code","status":["done"]}]}`,
	"v99 reshapes step":     `{"schemaVersion":99,"feature":"delta","currentStep":{"name":"code"},"steps":{}}`,
	"v99 reshapes routes":   `{"schemaVersion":99,"feature":"delta","currentStep":"code","steps":{},"modelRoutes":[]}`,
}

// TestFutureVersionNeverBricksLoad is the direct statement of the guarantee the
// whole design rests on: a file from a newer Centinela LOADS, whatever it did
// to the schema, so the cascade above is unreachable.
func TestFutureVersionNeverBricksLoad(t *testing.T) {
	for name, body := range futureFiles {
		t.Run(name, func(t *testing.T) {
			stateRepo(t)
			writeRawState(t, "delta", body)
			wf, err := Load("delta")
			if err != nil {
				t.Fatalf("a future-version file must never fail to load: %v", err)
			}
			if !wf.Unmodellable() {
				t.Fatal("a file this binary cannot unmarshal must be marked unmodellable")
			}
			if wf.Feature != "delta" {
				t.Fatalf("Feature = %q, want the salvaged \"delta\"", wf.Feature)
			}
			if wf.CurrentStep == "" || wf.CurrentStep == "done" {
				t.Fatalf("CurrentStep = %q — an inactive step drops the file from "+
					"ActiveWorkflows, which is the cascade this test exists to stop",
					wf.CurrentStep)
			}
			if err := Save(wf); err == nil {
				t.Fatal("saving over a file this binary cannot model must be refused")
			}
			after, _ := os.ReadFile(FilePath("delta"))
			if string(after) != body {
				t.Fatalf("a refused file must be byte-identical, got %q", after)
			}
		})
	}
}

// The anti-bricking invariant asserted where it actually lives, rather than
// through the prewrite hook: the active set must still contain the feature.
func TestActiveWorkflowsKeepsFutureVersionFile(t *testing.T) {
	for name, body := range futureFiles {
		t.Run(name, func(t *testing.T) {
			dir := stateRepo(t)
			writeRawState(t, "delta", body)
			wfs := ActiveWorkflows(dir + "/" + WorkflowDir)
			if len(wfs) != 1 {
				t.Fatalf("ActiveWorkflows = %d workflows, want 1 — an empty active "+
					"set is what blocks every governed write", len(wfs))
			}
			if wfs[0].Feature != "delta" {
				t.Fatalf("Feature = %q, want \"delta\"", wfs[0].Feature)
			}
		})
	}
}
