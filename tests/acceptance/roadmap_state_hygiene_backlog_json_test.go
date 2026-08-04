// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"encoding/json"
	"testing"
)

// Scenario: --json emits the machine shape for every finding
func TestRsh_BacklogJSONEmitsMachineShape(t *testing.T) {
	bin := buildCent(t)
	dir := rshBacklogFixture(t, 40, 2)
	out, code := runCent(t, bin, dir, "roadmap", "backlog", "--json")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	var payload struct {
		ThresholdDays int `json:"threshold_days"`
		Total         int `json:"total"`
		Stale         int `json:"stale"`
		Findings      []struct {
			Slug       string `json:"slug"`
			Summary    string `json:"summary"`
			DeferredAt string `json:"deferredAt"`
			AgeDays    *int   `json:"ageDays"`
			Stale      bool   `json:"stale"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout must be valid JSON: %v\n%s", err, out)
	}
	if payload.Total != 2 || len(payload.Findings) != 2 {
		t.Fatalf("want 2 findings, got %+v", payload)
	}
	for _, f := range payload.Findings {
		if f.Slug == "" || f.AgeDays == nil {
			t.Fatalf("finding must carry slug/deferredAt/ageDays: %+v", f)
		}
	}
}

// Scenario: an empty Backlog is an explicit empty state, not an error
func TestRsh_EmptyBacklogIsAnExplicitEmptyState(t *testing.T) {
	bin := buildCent(t)
	dir := acceptanceDir(t, `{"phases":[{"name":"P","features":[{"name":"a"}]}]}`)
	out, code := runCent(t, bin, dir, "roadmap", "backlog", "--stale")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	containsAll(t, out, "No deferred findings")
}
