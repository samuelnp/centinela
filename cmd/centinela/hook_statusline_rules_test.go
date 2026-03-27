package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestStatuslineRulesPlanAndTests(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	cfg := &config.Config{}
	wf := workflow.New("alpha")
	block, _ := statusBlockAndNext(wf, cfg)
	if block != "MISSING_FEATURE_BRIEF" {
		t.Fatalf("expected missing brief, got %s", block)
	}
	mkdir(t, "docs/features")
	write(t, "docs/features/alpha.md", "x")
	block, _ = statusBlockAndNext(wf, cfg)
	if block != "MISSING_PLAN" {
		t.Fatalf("expected missing plan, got %s", block)
	}
	mkdir(t, "docs/plans")
	write(t, "docs/plans/alpha.md", "x")
	block, _ = statusBlockAndNext(wf, cfg)
	if block != "MISSING_SPEC" {
		t.Fatalf("expected missing spec, got %s", block)
	}
	mkdir(t, "specs")
	write(t, "specs/alpha.feature", "Feature: x")
	block, _ = statusBlockAndNext(wf, cfg)
	if block != "none" {
		t.Fatalf("expected no plan block, got %s", block)
	}
	wf.CurrentStep = "tests"
	block, _ = statusBlockAndNext(wf, cfg)
	if block != "MISSING_EDGE_CASES" {
		t.Fatalf("expected missing edge cases, got %s", block)
	}
	wf.CurrentStep = "code"
	_, next := statusBlockAndNext(wf, cfg)
	if next != "implement-code" {
		t.Fatalf("expected code next action, got %s", next)
	}
}

func TestStatuslineRulesValidate(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	wf := workflow.New("alpha")
	wf.CurrentStep = "validate"
	cfg := &config.Config{Gates: config.GatesConfig{ProductionReadinessEnabled: true}}
	block, _ := statusBlockAndNext(wf, cfg)
	if block != "MISSING_GATEKEEPER" {
		t.Fatalf("expected missing gatekeeper, got %s", block)
	}
	mkdir(t, ".workflow")
	write(t, ".workflow/alpha-gatekeeper.md", "ok")
	block, _ = statusBlockAndNext(wf, cfg)
	if block != "MISSING_PROD_READINESS" {
		t.Fatalf("expected missing production readiness, got %s", block)
	}
	write(t, ".workflow/alpha-production-readiness.md", "**Status:** BLOCKING")
	block, _ = statusBlockAndNext(wf, cfg)
	if block != "PROD_BLOCKING" {
		t.Fatalf("expected prod blocking, got %s", block)
	}
}

func TestStatuslineRulesDocs(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	wf := workflow.New("alpha")
	wf.CurrentStep = "docs"
	cfg := &config.Config{}
	mkdir(t, ".workflow")
	workflow.Save(wf) //nolint:errcheck
	block, _ := statusBlockAndNext(wf, cfg)
	if block != "MISSING_DOCS_OUTPUT" {
		t.Fatalf("expected missing docs output, got %s", block)
	}
	mkdir(t, "docs/project-docs")
	write(t, "docs/project-docs/index.html", "<html></html>")
	block, _ = statusBlockAndNext(wf, cfg)
	if block != "MISSING_DOCS_EVIDENCE" {
		t.Fatalf("expected missing docs evidence, got %s", block)
	}
}
