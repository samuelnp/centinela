package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
)

var errSentinel = errors.New("strict orchestration evidence failed for \"plan\": missing: x")

func mustSave(t *testing.T, wf *Workflow) {
	t.Helper()
	if err := Save(wf); err != nil {
		t.Fatalf("save %s: %v", wf.Feature, err)
	}
}

// legacyWorkflow builds an in-flight workflow that predates planner-v1.
func legacyWorkflow(feature string) *Workflow {
	wf := New(feature)
	wf.PlanContract = ""
	return wf
}

// seedPlanDocs creates the brief + plan a plan-step evidence snapshot must list.
func seedPlanDocs(t *testing.T, feature string) {
	t.Helper()
	for _, dir := range []string{"docs/features", "docs/plans"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	write(t, filepath.Join("docs/features", feature+".md"), "brief")
	write(t, filepath.Join("docs/plans", feature+".md"), "plan")
}

// writePlanEvidence drops a schema-valid .md/.json pair for role.
func writePlanEvidence(t *testing.T, feature string, role orchestration.Role, handoff string) {
	t.Helper()
	write(t, orchestration.MarkdownPath(feature, role), "# report")
	body, err := json.Marshal(orchestration.Evidence{
		Feature:     feature,
		Step:        "plan",
		Role:        string(role),
		Status:      "done",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Inputs:      orchestration.RequiredPlanInputs(feature),
		Outputs:     []string{filepath.ToSlash(filepath.Join("docs/plans", feature+".md"))},
		EdgeCases:   []string{"an edge case"},
		HandoffTo:   handoff,
	})
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}
	write(t, orchestration.JSONPath(feature, role), string(body))
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
