package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// setupRoutingRepo writes a workflow + centinela.toml and runs the hook,
// returning its stdout. Exercises orchestrationRouting's config→domain mapping.
func setupRoutingRepo(t *testing.T, step, toml string) string {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) })       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	if toml != "" {
		os.WriteFile("centinela.toml", []byte(toml), 0644) //nolint:errcheck
	}
	wf := workflow.New("f")
	wf.CurrentStep = step
	workflow.Save(wf) //nolint:errcheck
	return captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookOrchestration(nil, nil) }) //nolint:errcheck
	})
}

// AC1: model_map tier remap flows through orchestrationRouting → directive.
func TestRunHookOrchestration_ModelMapRemap(t *testing.T) {
	out := setupRoutingRepo(t, "plan", "[orchestration.model_map.reasoning]\nopencode = \"moonshotai/kimi-k2\"\n")
	if !strings.Contains(out, "model: moonshotai/kimi-k2 (opencode)") {
		t.Fatalf("expected kimi for opencode reasoning, got: %s", out)
	}
	// AC7: codex column never leaks the opencode ID — shows the tier name.
	if !strings.Contains(out, "model: reasoning (codex)") {
		t.Fatalf("expected codex rule-4 tier name, got: %s", out)
	}
}

// AC2: role override flows through orchestrationRouting → directive.
func TestRunHookOrchestration_RoleOverride(t *testing.T) {
	out := setupRoutingRepo(t, "code", "[orchestration.models]\nsenior-engineer = { opencode = \"deepseek/deepseek-coder\" }\n")
	if !strings.Contains(out, "senior-engineer (model: claude-opus-4-7 (claude)") {
		t.Fatalf("expected claude default for senior-engineer, got: %s", out)
	}
	if !strings.Contains(out, "model: deepseek/deepseek-coder (opencode)") {
		t.Fatalf("expected override for opencode, got: %s", out)
	}
}

// AC6: absent tables → built-in defaults for all runners.
func TestRunHookOrchestration_AbsentTablesDefault(t *testing.T) {
	out := setupRoutingRepo(t, "plan", "")
	if !strings.Contains(out, "model: claude-opus-4-7 (claude)") {
		t.Fatalf("expected default claude reasoning ID, got: %s", out)
	}
}

// Back-compat: a plain tier string flows through orchestrationRouting's tier loop.
func TestRunHookOrchestration_PlainTierString(t *testing.T) {
	out := setupRoutingRepo(t, "plan", "[orchestration.models]\nbig-thinker = \"fast\"\n")
	if !strings.Contains(out, "big-thinker (model: claude-haiku-4-5-20251001 (claude)") {
		t.Fatalf("expected big-thinker remapped to fast tier, got: %s", out)
	}
}

// Config error path: a malformed config falls back to defaults (zero-config safe).
func TestRunHookOrchestration_ConfigErrorFallsBack(t *testing.T) {
	out := setupRoutingRepo(t, "plan", "[orchestration.model_map.turbo]\nopencode = \"x\"\n")
	if !strings.Contains(out, "big-thinker (model: claude-opus-4-7 (claude)") {
		t.Fatalf("expected fallback defaults on config error, got: %s", out)
	}
}
