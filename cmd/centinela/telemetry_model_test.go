package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// cfgDriver returns a config whose [orchestration] driver_model is set.
func cfgDriver(m string) *config.Config {
	c := &config.Config{}
	c.Orchestration.DriverModel = m
	return c
}

// TestResolveEmitModelPrefersWorkflow — a workflow's pinned DriverModel wins over
// the config fallback.
func TestResolveEmitModelPrefersWorkflow(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	wf := &workflow.Workflow{DriverModel: "claude-sonnet-4-6"}
	if got := resolveEmitModel(wf, cfgDriver("claude-opus-4-7")); got != "claude-sonnet-4-6" {
		t.Fatalf("workflow model should win, got %q", got)
	}
}

// TestResolveEmitModelFallsBackToConfig — no workflow model → config/env fallback.
func TestResolveEmitModelFallsBackToConfig(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	if got := resolveEmitModel(nil, cfgDriver("claude-opus-4-7")); got != "claude-opus-4-7" {
		t.Fatalf("nil wf should fall back to config, got %q", got)
	}
	wf := &workflow.Workflow{DriverModel: ""}
	if got := resolveEmitModel(wf, cfgDriver("claude-opus-4-7")); got != "claude-opus-4-7" {
		t.Fatalf("empty wf model should fall back to config, got %q", got)
	}
}

// TestResolveEmitModelEmpty — neither workflow nor config nor env → "".
func TestResolveEmitModelEmpty(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	if got := resolveEmitModel(nil, &config.Config{}); got != "" {
		t.Fatalf("no source should yield empty, got %q", got)
	}
}

// TestResolveEmitModelFromPicksFirstActive — the first workflow with a DriverModel
// wins; later ones and nils are skipped.
func TestResolveEmitModelFromPicksFirstActive(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	wfs := []*workflow.Workflow{
		nil,
		{DriverModel: ""},
		{DriverModel: "claude-haiku-4-5"},
		{DriverModel: "claude-opus-4-7"},
	}
	if got := resolveEmitModelFrom(wfs, cfgDriver("cfg")); got != "claude-haiku-4-5" {
		t.Fatalf("first active model should win, got %q", got)
	}
}

// TestResolveEmitModelFromFallsBack — no active workflow → config/env fallback,
// then empty.
func TestResolveEmitModelFromFallsBack(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	wfs := []*workflow.Workflow{nil, {DriverModel: ""}}
	if got := resolveEmitModelFrom(wfs, cfgDriver("cfg-model")); got != "cfg-model" {
		t.Fatalf("should fall back to config, got %q", got)
	}
	if got := resolveEmitModelFrom(nil, &config.Config{}); got != "" {
		t.Fatalf("nothing configured should yield empty, got %q", got)
	}
}
