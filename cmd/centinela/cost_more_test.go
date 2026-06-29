package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestRunCostJSON(t *testing.T) {
	seedCostRepo(t)
	costJSON = true
	defer func() { costJSON = false }()
	out := capture(t, func() error { return runCost(nil, nil) })
	if !strings.Contains(out, `"features"`) || !strings.Contains(out, `"over": true`) {
		t.Fatalf("expected JSON report with over flag, got %q", out)
	}
}

func TestEmitCostWarningNotOverIsSilent(t *testing.T) {
	seedCostRepo(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Cost.StepTokenBudget = 1_000_000 // generous → nothing over
	cfg.Cost.FeatureTokenBudget = 1_000_000
	out := capture(t, func() error { emitCostWarning(cfg); return nil })
	if strings.TrimSpace(out) != "" {
		t.Fatalf("within budget should be silent, got %q", out)
	}
}

func TestActiveWorkflowNoneActive(t *testing.T) {
	t.Chdir(t.TempDir())
	if wf := activeWorkflow(mustGetwd()); wf != nil {
		t.Fatalf("empty repo should have no active workflow, got %+v", wf)
	}
}

func TestEmitCostWarningNilConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	out := capture(t, func() error { emitCostWarning(nil); return nil })
	if strings.TrimSpace(out) != "" {
		t.Fatalf("nil cfg should be silent, got %q", out)
	}
}
