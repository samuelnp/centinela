package gatereport

import (
	"strings"
	"testing"
)

// A verifier in a worktree MUST build its own binary (the installed one lags
// the branch), so any path to a centinela binary is an honest argv.
func TestAssessAcceptsAnyPathToACentinelaBinary(t *testing.T) {
	for _, argv := range []string{
		`["/tmp/centinela-dogfood","validate"]`,
		`["./bin/centinela","validate"]`,
		`["centinela","validate"]`,
	} {
		report := blockReport(`{"revision":"a","treeDigest":"b","commands":[{"argv":` + argv + `,"exitCode":0}]}`)
		if err := Assess(report); err != nil {
			t.Fatalf("argv %s must be accepted: %v", argv, err)
		}
	}
}

// ...but an arbitrarily named scratch binary is not a centinela run.
func TestAssessRejectsNonCentinelaBinary(t *testing.T) {
	for _, argv := range []string{
		`["/tmp/cent-verify","validate"]`,
		`["make","validate"]`,
		`["centinela","validate","--fast"]`,
	} {
		report := blockReport(`{"revision":"a","treeDigest":"b","commands":[{"argv":` + argv + `,"exitCode":0}]}`)
		if err := Assess(report); err == nil {
			t.Fatalf("argv %s must be refused", argv)
		}
	}
}

// The scaffolded stub's exact shape: CRITICAL + empty commands must report
// BOTH, so the operator is told the report is ungrounded, not just refuted.
func TestAssessCriticalStubNamesTheGroundingFailure(t *testing.T) {
	stub := "**Status:** CRITICAL\n\n```json centinela:verification\n" +
		`{"revision":"","treeDigest":"","commands":[]}` + "\n```\n"
	err := Assess(stub)
	if err == nil {
		t.Fatal("a scaffolded stub must never be admissible")
	}
	for _, want := range []string{"CRITICAL", "no commands-run record", "docs/architecture/gatekeeper-prompt.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("stub error missing %q: %v", want, err)
		}
	}
}

func TestAssessValidateNonZeroExitBlocks(t *testing.T) {
	assertNoRecord(t, blockReport(
		`{"revision":"a","treeDigest":"b","commands":[{"argv":["centinela","validate"],"exitCode":1}]}`))
}

func TestAssessEmptyArgvEntryBlocks(t *testing.T) {
	err := Assess(blockReport(`{"revision":"a","treeDigest":"b","commands":[{"argv":[],"exitCode":0}]}`))
	if err == nil || !strings.Contains(err.Error(), "empty argv") {
		t.Fatalf("want empty-argv block, got %v", err)
	}
}

func TestAssessMissingStampBlocks(t *testing.T) {
	for _, body := range []string{
		`{"treeDigest":"b","commands":[{"argv":["centinela","validate"],"exitCode":0}]}`,
		`{"revision":"a","commands":[{"argv":["centinela","validate"],"exitCode":0}]}`,
	} {
		err := Assess(blockReport(body))
		if err == nil || !strings.Contains(err.Error(), "artifact stamp") {
			t.Fatalf("want stamp remedy for %q, got %v", body, err)
		}
	}
}

func TestFirstFindingFallsBackWhenAbsent(t *testing.T) {
	if got := FirstFinding("**Status:** CRITICAL\n"); !strings.Contains(got, "Findings section") {
		t.Fatalf("want placeholder, got %q", got)
	}
}

func TestFirstFindingSkipsNonFindingSections(t *testing.T) {
	report := "#### Inputs Read\n- the diff\n\n#### Findings\n- **the real one**\n"
	if got := FirstFinding(report); got != "the real one" {
		t.Fatalf("FirstFinding = %q", got)
	}
}
