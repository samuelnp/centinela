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

// separators lets a legend line ("SAFE | WARNING | CRITICAL") resolve to its
// leading token without treating the legend as three candidate verdicts.
var separators = strings.NewReplacer("|", " ", ",", " ")

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

// firstVerdict matches the FIRST token of s and nothing else. Scanning further
// would fail open on the most natural English hedge a failing verifier writes:
// "NOT SAFE" and "not safe" must be unparseable (⇒ blocked), never SAFE.
func firstVerdict(s string) string {
	fields := strings.Fields(separators.Replace(strings.TrimLeft(s, "*_` \t")))
	if len(fields) == 0 {
		return ""
	}
	token := strings.Trim(fields[0], "*_`.,;:")
	for _, v := range knownVerdicts {
		if token == v {
			return v
		}
	}
	return ""
}
