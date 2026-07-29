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
