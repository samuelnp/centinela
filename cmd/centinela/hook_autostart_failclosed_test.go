package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// autostartPrompt is a prompt autostart.ShouldStart accepts, so these tests
// exercise the config branch rather than the intent filter.
const autostartPrompt = `{"prompt":"please add release diagnostics"}`

// brokenTOMLPinningStrict is the exact shape F2 was found with: the profile is
// pinned to strict, and an unrelated typo makes the whole file unloadable.
const brokenTOMLPinningStrict = "[workflow]\nenforcement_profile = \"strict\"\nuse_worktrees = tru\n"

// autostartTree lays a startable existing-project tree with the given config.
func autostartTree(t *testing.T, toml string) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	if toml != "" {
		os.WriteFile("centinela.toml", []byte(toml), 0644) //nolint:errcheck
	}
}

// startedWorkflows returns every workflow state file autostart wrote.
func startedWorkflows(t *testing.T) []*workflow.Workflow {
	t.Helper()
	return workflow.ActiveWorkflows(workflow.WorkflowDir)
}

// TestAutostartFailsClosedOnUnparseableConfig: an unreadable centinela.toml
// makes the profile unknowable, so autostart declines and names the failure. It
// must NEVER pin a guided contract on a config it could not read — the pin is
// written once and for life, so fixing the TOML later cannot undo it.
func TestAutostartFailsClosedOnUnparseableConfig(t *testing.T) {
	autostartTree(t, brokenTOMLPinningStrict)
	withStdin(t, autostartPrompt, func() {
		out := captureStdout(t, func() {
			if err := runHookAutostart(nil, nil); err != nil {
				t.Fatalf("the hook must never break the session: %v", err)
			}
		})
		if !strings.Contains(out, "autostart declined") || !strings.Contains(out, "centinela.toml") {
			t.Fatalf("output must decline and name the parse failure, got %q", out)
		}
	})
	if wfs := startedWorkflows(t); len(wfs) != 0 {
		t.Fatalf("no workflow may be pinned on an unreadable config, got %+v", wfs)
	}
}

// TestAutostartHonoursStrictOnceTheConfigParses is the other half: the same
// pinned-strict config, minus the typo, must produce a STRICT workflow — never
// the guided tail. This is what silently broke while the error was discarded.
func TestAutostartHonoursStrictOnceTheConfigParses(t *testing.T) {
	autostartTree(t, "[workflow]\nenforcement_profile = \"strict\"\n")
	withStdin(t, autostartPrompt, func() {
		_ = captureStdout(t, func() { _ = runHookAutostart(nil, nil) })
	})
	wfs := startedWorkflows(t)
	if len(wfs) != 1 {
		t.Fatalf("expected exactly one autostarted workflow, got %+v", wfs)
	}
	if wfs[0].OrchestrationMode != workflow.StrictOrchestrationMode {
		t.Fatalf("strict must still require the orchestration evidence bundle, got %q",
			wfs[0].OrchestrationMode)
	}
}

// TestAutostartStillStartsOnAZeroConfigTree guards the happy path the fix must
// not have broken: no centinela.toml at all is not a read FAILURE.
func TestAutostartStillStartsOnAZeroConfigTree(t *testing.T) {
	autostartTree(t, "")
	withStdin(t, autostartPrompt, func() {
		_ = captureStdout(t, func() { _ = runHookAutostart(nil, nil) })
	})
	wfs := startedWorkflows(t)
	if len(wfs) != 1 || wfs[0].ProfileContract != workflow.ProfileContractGuidedDefault {
		t.Fatalf("a zero-config tree must still autostart under the guided default, got %+v", wfs)
	}
}
