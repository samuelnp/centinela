package delivery

import "testing"

func TestGatekeeperVerdictReadsStatusLineOnly(t *testing.T) {
	// Regression: prose mentioning "warnings" must NOT override a SAFE verdict.
	report := "### Gatekeeper Report\n**Status:** SAFE\n\n- import_graph shows non-failing warnings; not a regression\n"
	if got := gatekeeperVerdict(report); got != "SAFE" {
		t.Fatalf("prose 'warnings' skewed verdict: got %q, want SAFE", got)
	}
}

func TestGatekeeperVerdictTokens(t *testing.T) {
	cases := map[string]string{
		"**Status:** SAFE":                    "SAFE",
		"**Status:** WARNING":                 "WARNING",
		"**Status:** UNSAFE":                  "UNSAFE", // not mis-read as SAFE
		"Status: BLOCKING":                    "BLOCKING",
		"**Status:** SAFE | WARNING | BLOCK":  "SAFE", // first token wins (accidental legend)
		"no status line here, all is warning": "",     // not substring-matched
		"":                                    "",
	}
	for report, want := range cases {
		if got := gatekeeperVerdict(report); got != want {
			t.Fatalf("gatekeeperVerdict(%q) = %q, want %q", report, got, want)
		}
	}
}
