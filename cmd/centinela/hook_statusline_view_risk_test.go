package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestBuildStatusLineViewRiskWarnAndLegacyMode(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	mkdir(t, ".workflow")
	write(t, "centinela.toml", "[gates]\nproduction_readiness = true\n")
	write(t, ".workflow/alpha-production-readiness.md", "**Status:** WARNING")
	wf := workflow.New("alpha")
	wf.OrchestrationMode = ""
	v := buildStatusLineView([]*workflow.Workflow{wf})
	out := strings.Join(v.Secondary, " ")
	if !strings.Contains(out, "MODE:legacy") || !strings.Contains(out, "RISK:warn") {
		t.Fatalf("expected legacy mode and warning risk, got: %s", out)
	}
}
