package gatereport

import "testing"

func TestVerdictTokens(t *testing.T) {
	cases := map[string]string{
		"**Status:** SAFE":                      "SAFE",
		"**Status:** WARNING":                   "WARNING",
		"**Status:** CRITICAL":                  "CRITICAL",
		"**Status:** BLOCKING":                  "BLOCKING",
		"**Status:** UNSAFE":                    "UNSAFE",
		"Status: BLOCKING":                      "BLOCKING",
		"## Status: SAFE":                       "SAFE",
		"**Status:** safe":                      "SAFE",
		"**Status:** SAFE | WARNING | CRITICAL": "SAFE",
		"**Status:** SAFE.":                     "SAFE",
		"**Status:** **CRITICAL**":              "CRITICAL",
		// The natural English hedge a failing verifier writes MUST NOT read as
		// a pass: only the FIRST token is considered.
		"**Status:** NOT SAFE":                "",
		"**Status:** not safe":                "",
		"**Status:** probably SAFE":           "",
		"**Status:** mostly fine I think":     "",
		"no status line here, all is warning": "",
		"":                                    "",
	}
	for report, want := range cases {
		if got := Verdict(report); got != want {
			t.Fatalf("Verdict(%q) = %q, want %q", report, got, want)
		}
	}
}

func TestVerdictReadsStatusLineOnly(t *testing.T) {
	// Regression (PR #59): prose mentioning "warnings" must NOT override SAFE.
	report := "### Report\n**Status:** SAFE\n\n- import_graph shows non-failing warnings\n"
	if got := Verdict(report); got != VerdictSafe {
		t.Fatalf("prose skewed verdict: got %q, want SAFE", got)
	}
}

func TestNormalizeAliases(t *testing.T) {
	cases := map[string]string{
		"BLOCKING": VerdictCritical,
		"UNSAFE":   VerdictCritical,
		"CRITICAL": VerdictCritical,
		"SAFE":     VerdictSafe,
		"WARNING":  VerdictWarning,
		"":         "",
		"MAYBE":    "",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Fatalf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
