package roadmap

import (
	"bytes"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// seedStatus writes a workflow state so FeatureStatus(feature) resolves to step.
func seedStatus(t *testing.T, feature, step string) {
	t.Helper()
	wf := workflow.New(feature)
	wf.CurrentStep = step
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save workflow %s: %v", feature, err)
	}
}

// TestRemove_RefusesInProgressOrDone rejects a non-planned feature, naming status.
func TestRemove_RefusesInProgressOrDone(t *testing.T) {
	for _, tc := range []struct{ step, status string }{
		{"code", "in-progress"}, {"done", "done"},
	} {
		t.Run(tc.status, func(t *testing.T) {
			crudChdir(t, crudBody)
			seedStatus(t, "lonely-feature", tc.step)
			before := crudBytes(t, RoadmapFile)
			err := Remove(RoadmapFile, "lonely-feature")
			if err == nil || !strings.Contains(err.Error(), tc.status) {
				t.Fatalf("expected %s refusal, got %v", tc.status, err)
			}
			if !bytes.Equal(before, crudBytes(t, RoadmapFile)) {
				t.Fatal("refused remove must be byte-identical")
			}
		})
	}
}

// TestRemove_RefusesWhenDepended names the dependent and writes nothing.
func TestRemove_RefusesWhenDepended(t *testing.T) {
	crudChdir(t, crudBody)
	before := crudBytes(t, RoadmapFile)
	err := Remove(RoadmapFile, "auth-service") // checkout-ui depends on it
	if err == nil || !strings.Contains(err.Error(), "checkout-ui") {
		t.Fatalf("expected dependent refusal naming checkout-ui, got %v", err)
	}
	if !bytes.Equal(before, crudBytes(t, RoadmapFile)) {
		t.Fatal("refused remove must be byte-identical")
	}
}

// TestRemove_RefusesWhenDraftDepends treats a draft as a real dependent.
func TestRemove_RefusesWhenDraftDepends(t *testing.T) {
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[` +
		`{"name":"auth-service"},` +
		`{"name":"draft-consumer","dependsOn":["auth-service"],"draft":true}]}]}`
	crudChdir(t, body)
	err := Remove(RoadmapFile, "auth-service")
	if err == nil || !strings.Contains(err.Error(), "draft-consumer") {
		t.Fatalf("a draft dependent must still block remove, got %v", err)
	}
}
