package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCov2PlanAdvisorConfigWarnFallback drives the hook's config-error warn
// branch: a corrupt centinela.toml must NOT break the host session.
func TestCov2PlanAdvisorConfigWarnFallback(t *testing.T) {
	writeCorruptConfig(t)
	out := captureStdout(t, func() {
		feedStdin(t, "", func() {
			if err := runHookPlanAdvisor(nil, nil); err != nil {
				t.Fatalf("plan-advisor must never break on config error: %v", err)
			}
		})
	})
	if !strings.Contains(out, "config warning") {
		t.Fatalf("expected a config warning, got: %s", out)
	}
}

// TestCov2OrchestrationConfigWarnAndSkipsNonStrict covers both the config-error
// warn branch and the non-strict workflow continue (a workflow not in strict
// orchestration mode emits no directive).
func TestCov2OrchestrationConfigWarnAndSkipsNonStrict(t *testing.T) {
	d := writeCorruptConfig(t)
	if err := os.MkdirAll(filepath.Join(d, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, ".workflow", "feat.json"),
		[]byte(`{"feature":"feat","currentStep":"code","stepOrder":["plan","code"],"steps":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		feedStdin(t, "", func() {
			if err := runHookOrchestration(nil, nil); err != nil {
				t.Fatalf("orchestration hook must not error: %v", err)
			}
		})
	})
	if !strings.Contains(out, "config warning") {
		t.Fatalf("expected config warning, got: %s", out)
	}
	if strings.Contains(out, "orchestrator only") {
		t.Fatalf("non-strict workflow must emit no directive, got: %s", out)
	}
}

// TestCov2PrewriteConfigWarnFallback drives the prewrite hook's config-error
// warn-to-stderr fallback while still resolving a non-blocking decision.
func TestCov2PrewriteConfigWarnFallback(t *testing.T) {
	writeCorruptConfig(t)
	feedStdin(t, `{"tool_input":{"file_path":"README.md"}}`, func() {
		if err := runHookPrewrite(nil, nil); err != nil {
			t.Fatalf("prewrite must not break the host session on config error: %v", err)
		}
	})
}
