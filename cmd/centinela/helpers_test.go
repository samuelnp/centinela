package main

import "testing"

func TestAgentFlags(t *testing.T) {
	if !isValidAgent("claude") || !isValidAgent("opencode") || !isValidAgent("both") {
		t.Fatal("expected valid agents")
	}
	if isValidAgent("nope") {
		t.Fatal("unexpected valid agent")
	}
	if !usesClaude("both") || usesClaude("opencode") {
		t.Fatal("usesClaude mismatch")
	}
	if !usesOpenCode("both") || usesOpenCode("claude") {
		t.Fatal("usesOpenCode mismatch")
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
