package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// echoSkips is an acceptance-classified command (it contains "acceptance" and
// "test") whose stdout is a cucumber summary reporting a skip.
const echoSkips = `echo "3 scenarios (1 skipped, 2 passed)" # acceptance test`

// End-to-end through the runner: the run FAILS under the default policy.
func TestRunValidateCommands_AcceptanceSkipsFailTheRun(t *testing.T) {
	cfg := &config.Config{}
	cfg.Validate.Commands = []string{echoSkips}
	cfg.Validate.AcceptanceSkipPolicy = config.AcceptanceSkipFail
	out := captureStdout(t, func() {
		if runValidateCommands(cfg) {
			t.Error("a reported acceptance skip must fail the run")
		}
	})
	if !strings.Contains(out, "1 skipped") {
		t.Fatalf("output must name the skipped count, got %q", out)
	}
}

// The warn policy prints the counts and lets the run pass.
func TestRunValidateCommands_WarnPolicyPassesTheRun(t *testing.T) {
	cfg := &config.Config{}
	cfg.Validate.Commands = []string{echoSkips}
	cfg.Validate.AcceptanceSkipPolicy = config.AcceptanceSkipWarn
	out := captureStdout(t, func() {
		if !runValidateCommands(cfg) {
			t.Error("the warn policy must not fail the run")
		}
	})
	if !strings.Contains(out, "1 skipped") || !strings.Contains(out, "⚠") {
		t.Fatalf("expected a warning naming the counts, got %q", out)
	}
}

// A non-acceptance command reporting skips stays green.
func TestRunValidateCommands_NonAcceptanceSkipsStayGreen(t *testing.T) {
	cfg := &config.Config{}
	cfg.Validate.Commands = []string{`echo "--- SKIP: TestA (0.00s)"`}
	cfg.Validate.AcceptanceSkipPolicy = config.AcceptanceSkipFail
	captureStdout(t, func() {
		if !runValidateCommands(cfg) {
			t.Error("a non-acceptance command must never be failed by skip detection")
		}
	})
}

// E10: zero configured commands short-circuits before any analysis.
func TestRunValidateCommands_NoCommandsShortCircuits(t *testing.T) {
	cfg := &config.Config{}
	out := captureStdout(t, func() {
		if !runValidateCommands(cfg) {
			t.Error("no configured commands must pass")
		}
	})
	if strings.TrimSpace(out) != "" {
		t.Fatalf("nothing should be printed for zero commands, got %q", out)
	}
}
