package main

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestRoadmapJSONDirective_NotExistVsInvalid(t *testing.T) {
	if got := roadmapJSONDirective(os.ErrNotExist); !strings.Contains(got, "required") {
		t.Fatalf("ErrNotExist branch missing 'required': %q", got)
	}
	if got := roadmapJSONDirective(errors.New("syntax error")); !strings.Contains(got, "invalid") {
		t.Fatalf("invalid branch missing 'invalid': %q", got)
	}
}

func TestRoadmapCommandError_Branches(t *testing.T) {
	if got := roadmapCommandError(os.ErrNotExist); !strings.Contains(got.Error(), "missing") {
		t.Fatalf("ErrNotExist branch missing 'missing': %v", got)
	}
	if got := roadmapCommandError(errors.New("syntax")); !strings.Contains(got.Error(), "invalid") {
		t.Fatalf("invalid branch missing 'invalid': %v", got)
	}
}
