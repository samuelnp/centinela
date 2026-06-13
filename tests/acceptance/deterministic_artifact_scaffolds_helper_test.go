package acceptance_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// dasEvidence is the on-disk shape we assert against (mirrors the validator's
// view of the JSON). Only the list fields matter for these scenarios.
type dasEvidence struct {
	Inputs    []string `json:"inputs"`
	Outputs   []string `json:"outputs"`
	EdgeCases []string `json:"edgeCases"`
}

// Shared harness for the deterministic-artifact-scaffolds acceptance suite.
// It mirrors what `centinela evidence init` does in cmd/centinela without
// importing package main: Skeleton + PlanInputs pre-fill, atomic write,
// companion. Each test chdirs into an isolated tempdir.

func dasChdir(t *testing.T, briefs ...string) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(o) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, b := range briefs {
		if err := os.WriteFile("docs/features/"+b, []byte("brief"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// dasInit replicates `centinela evidence init <feature> <role>`: skeleton with
// the plan-snapshot pre-fill applied, written atomically, plus its companion.
func dasInit(t *testing.T, feature string, role evidence.Role) {
	t.Helper()
	skel := evidence.Skeleton(feature, role, "1.0.0")
	if pre := evidence.PlanInputs(feature, role); pre != nil {
		skel.Inputs = pre
	}
	if err := evidence.WriteAtomic(feature, role, skel); err != nil {
		t.Fatal(err)
	}
	if err := evidence.WriteCompanion(feature, role, evidence.DefaultCompanionTemplate(feature, role)); err != nil {
		t.Fatal(err)
	}
}

func dasReadJSON(t *testing.T, feature string, role evidence.Role) (string, dasEvidence) {
	t.Helper()
	data, err := os.ReadFile(orchestration.JSONPath(feature, role))
	if err != nil {
		t.Fatal(err)
	}
	var e dasEvidence
	if err := json.Unmarshal(data, &e); err != nil {
		t.Fatal(err)
	}
	return string(data), e
}
