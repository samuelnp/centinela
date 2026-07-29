package gatereport

import (
	"strings"
	"testing"
)

const goodBody = `{"revision":"9f2c1ab","treeDigest":"sha256:4e7d",` +
	`"commands":[{"argv":["centinela","validate"],"exitCode":0}]}`

func TestAssessHappyPath(t *testing.T) {
	if err := Assess(blockReport(goodBody)); err != nil {
		t.Fatalf("grounded report rejected: %v", err)
	}
}

func TestAssessWarningIsAdmissible(t *testing.T) {
	report := strings.Replace(blockReport(goodBody), "SAFE", "WARNING", 1)
	if err := Assess(report); err != nil {
		t.Fatalf("WARNING must advance: %v", err)
	}
}

func TestAssessCriticalEchoesFinding(t *testing.T) {
	report := "**Status:** CRITICAL\n\n#### Findings\n- session tokens are never invalidated\n"
	err := Assess(report)
	if err == nil || !strings.Contains(err.Error(), "session tokens are never invalidated") {
		t.Fatalf("want echoed finding, got %v", err)
	}
	if !strings.Contains(err.Error(), "FRESH verifier") {
		t.Fatalf("want fresh-context remedy, got %v", err)
	}
}

func TestAssessAliasesBlockAsCritical(t *testing.T) {
	for _, alias := range []string{"BLOCKING", "UNSAFE"} {
		err := Assess("**Status:** " + alias + "\n")
		if err == nil || !strings.Contains(err.Error(), "CRITICAL") {
			t.Fatalf("%s must block as CRITICAL, got %v", alias, err)
		}
	}
}

func TestAssessMissingOrGarbledStatusBlocks(t *testing.T) {
	for _, report := range []string{"#### Findings\n- a thing\n", "**Status:** mostly fine\n"} {
		err := Assess(report)
		if err == nil || !strings.Contains(err.Error(), "missing or unparseable") {
			t.Fatalf("want unparseable block for %q, got %v", report, err)
		}
	}
}
