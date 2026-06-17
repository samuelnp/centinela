package workflow

import (
	"strings"
	"testing"
)

// Scenario 7: a genuinely missing state file reports absence.
func TestLoadMissingReportsAbsence(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := Load("ghost")
	if err == nil {
		t.Fatal("expected error for missing workflow file")
	}
	if !strings.Contains(err.Error(), "no workflow found") {
		t.Fatalf("missing file must report absence, got: %v", err)
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("absence error should name the feature, got: %v", err)
	}
}
