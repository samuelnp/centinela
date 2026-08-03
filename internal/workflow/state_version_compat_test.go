package workflow

import (
	"testing"
)

// legacyState is shaped like the 133 real, git-tracked, versionless
// .workflow/<feature>.json files in this repo (and like the fixtures 48 test
// files hand-author): no schemaVersion, with the fields added over time that
// are back-compat-by-absence.
const legacyState = `{
  "feature": "legacy",
  "startedAt": "2026-01-02T03:04:05Z",
  "currentStep": "validate",
  "steps": {"plan": {"status": "completed", "completedAt": "2026-01-02T04:00:00Z"},
            "validate": {"status": "in-progress", "completedAt": null}},
  "stepOrder": ["plan", "code", "tests", "validate", "docs"],
  "orchestrationMode": "subagent",
  "enforcementProfile": "strict",
  "archetype": "canonical",
  "driverModel": "claude-opus-5",
  "validateContract": "adversarial-verifier",
  "planContract": "unified-planner",
  "modelRoutes": {"senior-engineer": {"tier": "reasoning", "decidedAt": "2026-01-02T05:00:00Z"}}
}`

// A real legacy file must survive Load -> Save with every field intact. This is
// the compatibility canary for the whole tracked corpus: nothing about adding
// the version key may drop a field the operator's file already carries.
func TestLegacyStateRoundTripsWithEveryFieldIntact(t *testing.T) {
	stateRepo(t)
	writeRawState(t, "legacy", legacyState)

	before, err := Load("legacy")
	if err != nil {
		t.Fatalf("legacy load: %v", err)
	}
	if before.SchemaVersion != 1 {
		t.Fatalf("absent version must read as 1, got %d", before.SchemaVersion)
	}
	if err := Save(before); err != nil {
		t.Fatalf("legacy save: %v", err)
	}

	after, err := Load("legacy")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	checks := []struct {
		name      string
		got, want string
	}{
		{"feature", after.Feature, before.Feature},
		{"currentStep", after.CurrentStep, before.CurrentStep},
		{"orchestrationMode", after.OrchestrationMode, before.OrchestrationMode},
		{"enforcementProfile", after.EnforcementProfile, before.EnforcementProfile},
		{"archetype", after.Archetype, before.Archetype},
		{"driverModel", after.DriverModel, before.DriverModel},
		{"validateContract", after.ValidateContract, before.ValidateContract},
		{"planContract", after.PlanContract, before.PlanContract},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Fatalf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
	if len(after.StepOrder) != len(before.StepOrder) {
		t.Fatalf("stepOrder = %v, want %v", after.StepOrder, before.StepOrder)
	}
	if after.Steps["plan"].Status != "completed" || after.Steps["plan"].CompletedAt == nil {
		t.Fatalf("steps lost detail: %+v", after.Steps)
	}
	route, ok := after.ModelRoutes["senior-engineer"]
	if !ok || route.Tier != "reasoning" || route.DecidedAt == "" {
		t.Fatalf("model route lost: %+v", after.ModelRoutes)
	}
	if !after.StartedAt.Equal(before.StartedAt) {
		t.Fatalf("startedAt = %v, want %v", after.StartedAt, before.StartedAt)
	}
	// The upgrade is silent and one-way: the reloaded file now carries it.
	if after.SchemaVersion != SchemaVersion {
		t.Fatalf("save must stamp the version, got %d", after.SchemaVersion)
	}
}
