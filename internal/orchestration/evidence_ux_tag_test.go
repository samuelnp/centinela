package orchestration

import (
	"strings"
	"testing"
)

func TestNormalizeUXTagCutsAtFirstColon(t *testing.T) {
	cases := []struct{ entry, want string }{
		{"mobile-first", "mobile-first"},                               // bare token still works
		{"mobile-first: renders at 80x24", "mobile-first"},             // E8
		{"loading-state: spinner: with a 3s timeout", "loading-state"}, // E9: first colon
		{"error-state:", "error-state"},                                // E10: trailing colon
		{"MOBILE_FIRST : Renders", "mobile-first"},                     // E13
		{"Empty State: nothing yet", "empty-state"},                    // case + space
		{"motion_and_reduced_motion", "motion-and-reduced-motion"},     // underscores
		{"  responsive-layout  ", "responsive-layout"},                 // outer trim
		{":", ""},      // E11
		{": text", ""}, // E11
		{"", ""},       // E11
		{"   ", ""},    // E11
	}
	for _, tc := range cases {
		if got := normalizeUXTag(tc.entry); got != tc.want {
			t.Errorf("normalizeUXTag(%q) = %q, want %q", tc.entry, got, tc.want)
		}
	}
}

// Degenerate entries land in the tag set but satisfy nothing (AC12).
func TestDegenerateEntriesSatisfyNoRequiredTag(t *testing.T) {
	missing := missingUXTags([]string{":", ": text", "", "   "})
	if len(missing) != len(requiredUXTags) {
		t.Fatalf("degenerate entries matched a tag: missing=%v", missing)
	}
}

func TestValidateUXEvidenceAcceptsDescriptiveEntries(t *testing.T) {
	mobileFirst := true
	descriptive := make([]string, 0, len(requiredUXTags))
	for _, tag := range requiredUXTags {
		descriptive = append(descriptive, tag+": stated once, with detail")
	}
	if err := validateUXEvidence("x", RoleUXUISpecialist, descriptive, &mobileFirst); err != nil {
		t.Fatalf("all-descriptive evidence must pass, got %v", err)
	}
}

// Regression guard (AC10): the change only ADDS matches — bare tokens pass.
func TestValidateUXEvidenceStillAcceptsBareTokens(t *testing.T) {
	mobileFirst := true
	if err := validateUXEvidence("x", RoleUXUISpecialist, requiredUXTags, &mobileFirst); err != nil {
		t.Fatalf("bare tokens must still pass, got %v", err)
	}
}

// E12: a bare token and its descriptive twin dedup to one satisfied tag.
func TestValidateUXEvidenceDedupsBareAndDescriptiveTwin(t *testing.T) {
	mobileFirst := true
	entries := []string{"empty-state", "empty-state: nothing yet"}
	for _, tag := range requiredUXTags {
		if tag != "empty-state" {
			entries = append(entries, tag+": covered")
		}
	}
	if err := validateUXEvidence("x", RoleUXUISpecialist, entries, &mobileFirst); err != nil {
		t.Fatalf("dedup case must pass, got %v", err)
	}
}

// AC11: a genuinely absent tag is still named, and covered tags are not.
func TestValidateUXEvidenceStillNamesGenuinelyMissingTag(t *testing.T) {
	mobileFirst := true
	var entries []string
	for _, tag := range requiredUXTags {
		if tag != "error-state" {
			entries = append(entries, tag+": covered descriptively")
		}
	}
	err := validateUXEvidence("x", RoleUXUISpecialist, entries, &mobileFirst)
	if err == nil || !strings.Contains(err.Error(), "error-state") {
		t.Fatalf("expected error-state reported missing, got %v", err)
	}
	for _, tag := range requiredUXTags {
		if tag != "error-state" && strings.Contains(err.Error(), tag) {
			t.Errorf("covered tag %q must not be reported missing: %v", tag, err)
		}
	}
}
