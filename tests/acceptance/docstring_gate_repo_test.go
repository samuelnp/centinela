package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/docstring-gate.feature
//
// This repository's own shipped posture, and the senior-engineer prompt's
// doc-comment duty.

// Scenario: This repository ships the gate enforcing at fail severity
func TestDG_ShipsFailSeverity(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "centinela.toml"))
	if err != nil {
		t.Fatal(err)
	}
	toml := string(data)
	idx := strings.Index(toml, "[gates.docstring]")
	if idx < 0 {
		t.Fatal("centinela.toml must register [gates.docstring]")
	}
	block := toml[idx:]
	if end := strings.Index(block[1:], "\n["); end >= 0 {
		block = block[:end+1]
	}
	mustHave(t, block, "enabled  = true")
	mustHave(t, block, "severity = \"fail\"")
	mustNotContain(t, block, "scope =")
}

// Scenario: The senior-engineer prompt carries the doc-comment duty in both copies
//
// Byte-identical mirror + 130-line budget are already asserted generically
// by TestScaffoldArchitectureMirrorParity and
// TestPromoteOrchestrationAgents_LineBudget; this pins the content itself.
func TestDG_SeniorEngineerPromptStatesDuty(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "docs", "architecture", "senior-engineer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, string(data), "doc comment")
	mustContain(t, string(data), "docstring gate")
}
