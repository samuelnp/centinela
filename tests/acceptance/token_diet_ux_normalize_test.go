// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"testing"
)

// tdUXMissing writes a single-entry ux-ui-specialist edge case and returns the
// validation error's message (empty string on an unexpected pass).
func tdUXMissing(t *testing.T, entry string) string {
	t.Helper()
	err := tdWriteUXEvidence(t, "token-diet", []string{entry}, true)
	if err == nil {
		return ""
	}
	return err.Error()
}

// Scenario: Tag normalization cuts at the first colon
func TestTD_TagNormalizationCutsAtFirstColon(t *testing.T) {
	cases := []struct{ entry, tag string }{
		{"mobile-first", "mobile-first"},
		{"mobile-first: renders at 80x24", "mobile-first"},
		{"loading-state: spinner: with a 3s timeout", "loading-state"},
		{"error-state:", "error-state"},
		{"MOBILE_FIRST : Renders", "mobile-first"},
		{"Empty State: nothing yet", "empty-state"},
	}
	for _, tc := range cases {
		msg := tdUXMissing(t, tc.entry)
		if msg == "" {
			t.Fatalf("entry %q: expected 7 other tags still missing", tc.entry)
		}
		mustNotContain(t, msg, tc.tag)
	}
}

// Scenario: Degenerate entries match nothing and never panic
func TestTD_DegenerateEntriesMatchNothingAndNeverPanic(t *testing.T) {
	for _, entry := range []string{":", ": text", ""} {
		msg := tdUXMissing(t, entry)
		if msg == "" {
			t.Fatalf("entry %q: a degenerate entry must satisfy no required tag", entry)
		}
		for _, tag := range []string{"mobile-first", "visual-hierarchy", "typography-hierarchy",
			"responsive-layout", "loading-state", "empty-state", "error-state",
			"motion-and-reduced-motion"} {
			mustContain(t, msg, tag)
		}
	}
}
