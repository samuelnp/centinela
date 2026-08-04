package hookpolicy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// degradedFixture loads a real future-versioned state file, the only way to get
// an unmodellable workflow: the flag is set by Load, never by a setter.
func degradedFixture(t *testing.T, feature string) *workflow.Workflow {
	t.Helper()
	body := `{"schemaVersion":99,"feature":"` + feature + `","steps":42}`
	if err := os.WriteFile(workflow.FilePath(feature), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	wf, err := workflow.Load(feature)
	if err != nil {
		t.Fatalf("a future-version file must load: %v", err)
	}
	if !wf.Unmodellable() {
		t.Fatal("precondition: the fixture must be unmodellable")
	}
	return wf
}

func degradedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A state file this binary cannot model must not widen permissions for a
// feature it CAN model: returning Allow for the degraded one let a single
// 31-byte file disarm step gating across the whole repo.
func TestDegradedWorkflowDoesNotDisarmAnEnforceableOne(t *testing.T) {
	dir := degradedRepo(t)
	real := &workflow.Workflow{Feature: "real", CurrentStep: "plan"}

	for _, wfs := range [][]*workflow.Workflow{
		{degradedFixture(t, "junk"), real},
		{real, degradedFixture(t, "junk")},
	} {
		got := EvaluatePrewrite(filepath.Join(dir, "internal", "evil.go"), dir, &config.Config{}, wfs)
		if got.Allow {
			t.Fatal("a degraded state file must not allow a code write during another feature's plan step")
		}
		if got.Feature != "real" || got.Step != "plan" {
			t.Fatalf("refusal must name the enforceable workflow, got feature=%q step=%q", got.Feature, got.Step)
		}
	}
}

// With no enforceable step to appeal to, the write is REFUSED rather than
// waved through: passing would be a self-service bypass, since an agent can
// write its own state file. The refusal must be the stale-binary one, so the
// message names upgrading rather than "run centinela start".
func TestOnlyDegradedWorkflowsRefuseWithStaleBinary(t *testing.T) {
	dir := degradedRepo(t)
	wfs := []*workflow.Workflow{degradedFixture(t, "junk"), degradedFixture(t, "other")}
	got := EvaluatePrewrite(filepath.Join(dir, "internal", "evil.go"), dir, &config.Config{}, wfs)
	if got.Allow {
		t.Fatal("an all-unmodellable repo must not open the gate")
	}
	if !got.StaleBinary || got.NeedInit {
		t.Fatalf("must refuse as stale-binary, not as need-init, got %+v", got)
	}
}

// A done workflow is not enforceable, so a degraded one beside it is still the
// only ACTIVE workflow and the stale-binary refusal applies rather than the
// out-of-step one.
func TestDegradedBesideDoneWorkflowRefusesAsStaleBinary(t *testing.T) {
	dir := degradedRepo(t)
	done := &workflow.Workflow{Feature: "shipped", CurrentStep: "done"}
	wfs := []*workflow.Workflow{done, degradedFixture(t, "junk")}
	got := EvaluatePrewrite(filepath.Join(dir, "internal", "evil.go"), dir, &config.Config{}, wfs)
	if got.Allow || !got.StaleBinary {
		t.Fatalf("a lone degraded workflow must refuse as stale-binary, got %+v", got)
	}
}
