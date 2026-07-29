package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func seedStatusline(t *testing.T, report string) *config.Config {
	t.Helper()
	t.Chdir(t.TempDir())
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	wf := workflow.New("f")
	// Orchestration evidence is a separate gate; these cases isolate the
	// verifier-report classification.
	wf.OrchestrationMode = ""
	wf.CurrentStep = "validate"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	if report != "" {
		os.WriteFile(workflow.GatekeeperReportPath("f"), []byte(report), 0o644) //nolint:errcheck
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestValidateStatusBlockMissingReport(t *testing.T) {
	cfg := seedStatusline(t, "")
	if block, next := validateStatusBlock("f", cfg); block != "MISSING_GATEKEEPER" || next != "run-verifier" {
		t.Fatalf("got %q/%q", block, next)
	}
}

func TestValidateStatusBlockUngroundedReport(t *testing.T) {
	cfg := seedStatusline(t, "**Status:** SAFE\n")
	if block, next := validateStatusBlock("f", cfg); block != "MISSING_COMMANDS_RECORD" || next != "rerun-verifier" {
		t.Fatalf("got %q/%q", block, next)
	}
}

func TestValidateStatusBlockGroundedReportClears(t *testing.T) {
	cfg := seedStatusline(t, "**Status:** SAFE\n\n```json centinela:verification\n"+
		`{"revision":"a","treeDigest":"b","commands":[{"argv":["centinela","validate"],"exitCode":0}]}`+
		"\n```\n")
	if block, next := validateStatusBlock("f", cfg); block != "none" || next != "run-validate" {
		t.Fatalf("got %q/%q", block, next)
	}
}
