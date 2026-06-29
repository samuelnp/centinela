package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestAgentFlags(t *testing.T) {
	if !isValidAgent("claude") || !isValidAgent("opencode") || !isValidAgent("aider") || !isValidAgent("both") {
		t.Fatal("expected valid agents")
	}
	if isValidAgent("nope") {
		t.Fatal("unexpected valid agent")
	}
	both, err := setup.AgentsFor("both")
	if err != nil || strings.Join(both, ",") != "claude,opencode" {
		t.Fatalf("expected both to resolve to claude,opencode: %v %v", both, err)
	}
	if msg := invalidAgentError("nope").Error(); !strings.Contains(msg, "aider") {
		t.Fatalf("invalid agent error should list registered harnesses: %s", msg)
	}
}

func TestNextStep(t *testing.T) {
	if nextStep("plan") != "code" || nextStep("validate") != "docs" || nextStep("docs") != "done" {
		t.Fatal("nextStep mismatch")
	}
}

func TestRunCommand(t *testing.T) {
	ok, out := runCommand("printf hello")
	if !ok || out != "hello" {
		t.Fatalf("runCommand expected hello, got ok=%v out=%q", ok, out)
	}
	fail, _ := runCommand("false")
	if fail {
		t.Fatal("expected failing command result")
	}
}
