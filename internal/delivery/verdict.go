package delivery

import "strings"

// knownVerdicts are the gatekeeper verdict tokens. "UNSAFE" precedes "SAFE" so
// exact-word matching never confuses the two.
var knownVerdicts = []string{"UNSAFE", "BLOCKING", "WARNING", "SAFE"}

// gatekeeperVerdict extracts the verdict from the report's "Status:" line ONLY
// (e.g. "**Status:** SAFE"), returning the first verdict token on that line.
// Scanning just the Status line — not the whole report — stops prose mentions
// of "warning"/"unsafe" elsewhere (e.g. "non-failing warnings") from skewing
// the result. Returns "" when there is no Status line naming a known verdict.
func gatekeeperVerdict(report string) string {
	for _, ln := range strings.Split(report, "\n") {
		head := strings.ToUpper(strings.TrimLeft(strings.TrimSpace(ln), "*# "))
		if !strings.HasPrefix(head, "STATUS:") {
			continue
		}
		return firstVerdict(head[len("STATUS:"):])
	}
	return ""
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
