package roadmap

import (
	"strings"
	"testing"
)

// TestParseScores_Valid accepts a well-formed CSV.
func TestParseScores_Valid(t *testing.T) {
	s, err := ParseScores("9,9,8,7,9,9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Overall != 9 || s.AcceptanceCriteria != 9 {
		t.Errorf("unexpected scores: %+v", s)
	}
}

// TestParseScores_Boundaries tests boundary values 1 and 10.
func TestParseScores_Boundaries(t *testing.T) {
	if _, err := ParseScores("1,1,1,1,1,9"); err != nil {
		t.Errorf("boundary 1 should be valid: %v", err)
	}
	if _, err := ParseScores("10,10,10,10,10,10"); err != nil {
		t.Errorf("boundary 10 should be valid: %v", err)
	}
}

// TestParseScores_NoOverallMinimum: the self-graded threshold is deleted, so
// every in-range overall parses. The RANGE check is what still bites (below).
func TestParseScores_NoOverallMinimum(t *testing.T) {
	for _, csv := range []string{"9,9,8,7,9,1", "3,3,3,3,3,3", "9,9,8,7,9,8", "9,9,8,7,9,10"} {
		if _, err := ParseScores(csv); err != nil {
			t.Errorf("%q must be accepted with no minimum: %v", csv, err)
		}
	}
}

// TestParseScores_OutOfRange rejects 0, 11, -1 — on EVERY field including
// overall. Deleting the minimum must not have deleted the 1-10 range with it.
func TestParseScores_OutOfRange(t *testing.T) {
	for _, csv := range []string{"0,9,9,9,9,9", "11,9,9,9,9,9", "-1,9,9,9,9,9",
		"9,9,9,9,9,0", "9,9,9,9,9,11"} {
		_, err := ParseScores(csv)
		if err == nil {
			t.Errorf("out-of-range scores %q should be rejected", csv)
			continue
		}
		if !strings.Contains(err.Error(), "must be between 1 and 10") {
			t.Errorf("%q must fail as a RANGE fault naming the bounds, got %v", csv, err)
		}
	}
}

// TestParseScores_WrongCount rejects 5 or 7 values.
func TestParseScores_WrongCount(t *testing.T) {
	if _, err := ParseScores("9,9,9,9,9"); err == nil {
		t.Error("5 values must be rejected")
	}
	if _, err := ParseScores("9,9,9,9,9,9,9"); err == nil {
		t.Error("7 values must be rejected")
	}
}

// TestParseScores_NonNumeric rejects non-integer tokens.
func TestParseScores_NonNumeric(t *testing.T) {
	_, err := ParseScores("9,abc,9,9,9,9")
	if err == nil {
		t.Error("non-numeric token must be rejected")
	}
	if !strings.Contains(err.Error(), "six comma-separated integers") {
		t.Errorf("error message should describe format, got: %v", err)
	}
}

// TestParseScores_Empty rejects an empty string.
func TestParseScores_Empty(t *testing.T) {
	if _, err := ParseScores(""); err == nil {
		t.Error("empty scores must be rejected")
	}
}

// TestParseScores_LowOverallRecorded: a low overall is recorded verbatim, not
// coerced or rejected — the scores survive as a record, they just gate nothing.
func TestParseScores_LowOverallRecorded(t *testing.T) {
	s, err := ParseScores("3,3,3,3,3,3")
	if err != nil || s.Overall != 3 {
		t.Errorf("overall=3 must parse and be recorded; err=%v scores=%+v", err, s)
	}
}
