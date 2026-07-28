package gatereport

import (
	"strings"
	"testing"
)

// assertNoRecord asserts the operator-facing "no commands-run record" remedy,
// which the spec demands for every ungrounded shape.
func assertNoRecord(t *testing.T, report string) {
	t.Helper()
	err := Assess(report)
	if err == nil || !strings.Contains(err.Error(), "no commands-run record") {
		t.Fatalf("want no-commands-run-record block, got %v", err)
	}
	if !strings.Contains(err.Error(), "docs/architecture/gatekeeper-prompt.md") {
		t.Fatalf("error must name the remedy doc, got %v", err)
	}
}

func TestAssessAbsentBlockBlocks(t *testing.T) {
	assertNoRecord(t, "### Report\n**Status:** SAFE\n")
}

func TestAssessEmptyCommandsBlocks(t *testing.T) {
	assertNoRecord(t, blockReport(`{"revision":"a","treeDigest":"b","commands":[]}`))
}

func TestAssessWithoutPassingValidateBlocks(t *testing.T) {
	assertNoRecord(t, blockReport(
		`{"revision":"a","treeDigest":"b","commands":[{"argv":["go","test","./..."],"exitCode":0}]}`))
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
