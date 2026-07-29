package gatereport

import (
	"strings"
	"testing"
)

func TestStampedCreatesBlockWhenAbsent(t *testing.T) {
	out, err := Stamped("### Report\n**Status:** SAFE\n", "abc123", "sha256:dead")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, err := ParseVerification(out)
	if err != nil {
		t.Fatalf("created block does not parse: %v", err)
	}
	if v.Revision != "abc123" || v.TreeDigest != "sha256:dead" || len(v.Commands) != 0 {
		t.Fatalf("bad created block: %+v", v)
	}
	// Fail-closed: a stamped-but-unfilled report is still inadmissible.
	if err := Assess(out); err == nil {
		t.Fatal("stamping alone must not make a report admissible")
	}
}

func TestStampedUpdatesInPlaceAndPreservesCommands(t *testing.T) {
	commands := `[{"argv":["centinela","validate"],"exitCode":0,"durationMs":84210}]`
	report := blockReport(`{"revision":"old","treeDigest":"stale","commands":` + commands + `}`)
	out, err := Stamped(report, "new", "sha256:fresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, commands) {
		t.Fatalf("commands array not byte-identical:\n%s", out)
	}
	if strings.Contains(out, "old") || strings.Contains(out, "stale") {
		t.Fatalf("stale stamp survived:\n%s", out)
	}
	if !strings.HasPrefix(out, "### Report\n**Status:** SAFE\n") {
		t.Fatalf("report body outside the block was rewritten:\n%s", out)
	}
	if err := Assess(out); err != nil {
		t.Fatalf("restamped grounded report rejected: %v", err)
	}
}

func TestStampedIsIdempotent(t *testing.T) {
	first, _ := Stamped("body\n", "r", "d")
	second, err := Stamped(first, "r", "d")
	if err != nil || first != second {
		t.Fatalf("stamp not idempotent: %v\n%q\n%q", err, first, second)
	}
}

func TestStampedRejectsMalformedBlock(t *testing.T) {
	if _, err := Stamped(blockReport("{nope"), "r", "d"); err == nil {
		t.Fatal("want malformed-block error")
	}
}
