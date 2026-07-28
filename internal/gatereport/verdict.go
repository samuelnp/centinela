package gatereport

import "strings"

// Canonical verdicts (D5). BLOCKING and UNSAFE are retained as accepted
// aliases that normalize to CRITICAL: legacy reports emit them, and
// production-readiness still emits BLOCKING (out of scope for this feature).
const (
	VerdictSafe     = "SAFE"
	VerdictWarning  = "WARNING"
	VerdictCritical = "CRITICAL"
)

// knownVerdicts are the tokens accepted on a Status line. "UNSAFE" precedes
// "SAFE" so exact-word matching never confuses the two.
var knownVerdicts = []string{"UNSAFE", "CRITICAL", "BLOCKING", "WARNING", "SAFE"}

// Verdict extracts the RAW verdict token from the report's "Status:" line only
// (e.g. "**Status:** SAFE"). Returns "" when there is no Status line naming a
// known verdict. Callers that reason about severity use Normalize; the
// delivery composer surfaces the raw token verbatim.
func Verdict(report string) string {
	for _, ln := range strings.Split(report, "\n") {
		head := strings.ToUpper(strings.TrimLeft(strings.TrimSpace(ln), "*# "))
		if !strings.HasPrefix(head, "STATUS:") {
			continue
		}
		return firstVerdict(head[len("STATUS:"):])
	}
	return ""
}

// Normalize maps the accepted aliases onto the canonical vocabulary. An
// unknown or empty token normalizes to "" — never to a passing verdict.
func Normalize(verdict string) string {
	switch verdict {
	case "BLOCKING", "UNSAFE", VerdictCritical:
		return VerdictCritical
	case VerdictSafe, VerdictWarning:
		return verdict
	default:
		return ""
	}
}

// firstVerdict returns the first known verdict token appearing in s by word
// position (treating | and , as separators), or "" when none is present.
func firstVerdict(s string) string {
	for _, w := range strings.Fields(strings.NewReplacer("|", " ", ",", " ").Replace(s)) {
		for _, v := range knownVerdicts {
			if w == v {
				return v
			}
		}
	}
	return ""
}
