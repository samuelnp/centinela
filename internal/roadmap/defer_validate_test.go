package roadmap

import (
	"strings"
	"testing"
)

// TestValidateSlug_Valid ensures legal kebab-case slugs pass.
func TestValidateSlug_Valid(t *testing.T) {
	for _, s := range []string{"hook-timeout", "my-feature-123", "a", "abc"} {
		if err := validateSlug(s); err != nil {
			t.Errorf("validateSlug(%q) unexpected error: %v", s, err)
		}
	}
}

// TestValidateSlug_Invalid covers uppercase, spaces, unicode, path chars, empty.
func TestValidateSlug_Invalid(t *testing.T) {
	cases := []string{
		"",
		"Bad Slug",
		"bad slug!",
		"CamelCase",
		"../escape",
		"unicode-ñ",
		"-starts-with-dash",
		"ends-with-dash-",
	}
	for _, s := range cases {
		if err := validateSlug(s); err == nil {
			t.Errorf("validateSlug(%q) should fail but did not", s)
		}
	}
}

// TestValidateSummary covers empty and whitespace-only rejections.
func TestValidateSummary_Empty(t *testing.T) {
	if err := validateSummary(""); err == nil {
		t.Error("empty summary must be rejected")
	}
	if err := validateSummary("   "); err == nil {
		t.Error("whitespace-only summary must be rejected")
	}
	// Verify error text.
	err := validateSummary("")
	if !strings.Contains(err.Error(), "summary") {
		t.Errorf("error should mention 'summary', got: %v", err)
	}
}

// TestValidateSummary_Valid accepts a non-empty summary.
func TestValidateSummary_Valid(t *testing.T) {
	if err := validateSummary("some finding"); err != nil {
		t.Errorf("valid summary rejected: %v", err)
	}
}

// TestValidateNoCollision_Clean passes when slug is not present.
func TestValidateNoCollision_Clean(t *testing.T) {
	existing := map[string]string{"other": "Phase 0"}
	if err := validateNoCollision("new-slug", existing); err != nil {
		t.Errorf("unexpected collision error: %v", err)
	}
}

// TestValidateNoCollision_Collision rejects a duplicate slug.
func TestValidateNoCollision_Collision(t *testing.T) {
	existing := map[string]string{"dupe": "Backlog", "real": "Phase 1"}
	if err := validateNoCollision("dupe", existing); err == nil {
		t.Error("expected collision error")
	} else if !strings.Contains(err.Error(), "dupe") {
		t.Errorf("error should name slug, got: %v", err)
	}
}

// TestValidateNoCollision_NonBacklogPhase reports the correct phase name.
func TestValidateNoCollision_NonBacklogPhase(t *testing.T) {
	existing := map[string]string{"enforce-coverage-in-validate": "Phase 3"}
	err := validateNoCollision("enforce-coverage-in-validate", existing)
	if err == nil {
		t.Fatal("expected collision error for slug in non-Backlog phase")
	}
	if !strings.Contains(err.Error(), "Phase 3") {
		t.Errorf("error must name containing phase, got: %v", err)
	}
}
