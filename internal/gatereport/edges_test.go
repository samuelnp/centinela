package gatereport

import (
	"strings"
	"testing"
)

func TestStripBulletIgnoresNonBullets(t *testing.T) {
	for _, line := range []string{"plain prose", "", "1. numbered"} {
		if got := stripBullet(line); got != "" {
			t.Fatalf("stripBullet(%q) = %q, want empty", line, got)
		}
	}
	if got := stripBullet("* starred"); got != "starred" {
		t.Fatalf("stripBullet star form = %q", got)
	}
}

func TestFirstVerdictOnEmptyRemainder(t *testing.T) {
	// "**Status:**" with nothing after it must be unparseable, never a pass.
	if got := Verdict("**Status:**\n"); got != "" {
		t.Fatalf("bare Status line = %q, want empty", got)
	}
	if got := Verdict("**Status:** ***\n"); got != "" {
		t.Fatalf("punctuation-only Status = %q, want empty", got)
	}
}

func TestStampedAppendsToReportLackingTrailingNewline(t *testing.T) {
	out, err := Stamped("**Status:** SAFE", "rev", "dig")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "**Status:** SAFE\n") {
		t.Fatalf("missing newline before the appended block:\n%q", out)
	}
	v, err := ParseVerification(out)
	if err != nil || v.Revision != "rev" {
		t.Fatalf("appended block does not parse: %+v %v", v, err)
	}
}

func TestStampedOnEmptyReport(t *testing.T) {
	out, err := Stamped("", "rev", "dig")
	if err != nil {
		t.Fatal(err)
	}
	if v, err := ParseVerification(out); err != nil || v.TreeDigest != "dig" {
		t.Fatalf("empty report stamp failed: %+v %v", v, err)
	}
}
