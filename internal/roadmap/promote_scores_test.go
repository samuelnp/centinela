package roadmap

import (
	"strings"
	"testing"
)

// TestParseScores_Valid accepts a well-formed CSV with overall >= 9.
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

// TestParseScores_OverallThreshold tests overall=8 rejected, 9 accepted.
func TestParseScores_OverallThreshold(t *testing.T) {
	if _, err := ParseScores("9,9,8,7,9,8"); err == nil {
		t.Error("overall=8 must be rejected")
	}
	if _, err := ParseScores("9,9,8,7,9,9"); err != nil {
		t.Errorf("overall=9 must be accepted: %v", err)
	}
	if _, err := ParseScores("9,9,8,7,9,10"); err != nil {
		t.Errorf("overall=10 must be accepted: %v", err)
	}
}

// TestParseScores_OutOfRange rejects 0, 11, -1.
func TestParseScores_OutOfRange(t *testing.T) {
	for _, csv := range []string{"0,9,9,9,9,9", "11,9,9,9,9,9", "-1,9,9,9,9,9"} {
		if _, err := ParseScores(csv); err == nil {
			t.Errorf("out-of-range scores %q should be rejected", csv)
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

// TestParseScores_ExactlyNine tests the minimum overall passes.
func TestParseScores_ExactlyNine(t *testing.T) {
	s, err := ParseScores("9,9,9,9,9,9")
	if err != nil || s.Overall != 9 {
		t.Errorf("overall=9 must pass; err=%v scores=%+v", err, s)
	}
}
