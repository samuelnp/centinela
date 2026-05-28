package acceptance_test

import (
	"strings"
	"testing"
)

// proximityWindow is the maximum character distance between a forbidden
// pattern (python heredoc, jq, cat heredoc, etc.) and a `.workflow`
// reference that still counts as "the agent is using that tool to write
// orchestration evidence". 200 chars is large enough to catch
// multi-line examples and small enough to ignore unrelated mentions of
// `jq` in prose far from the `.workflow` path.
const proximityWindow = 200

func assertNoForbiddenNearWorkflow(t *testing.T, label, body string) {
	t.Helper()
	patterns := []struct {
		needle string
		near   string
	}{
		{"python3 -c", ".workflow"},
		{"python3 <<", ".workflow"},
		{"<<EOF", ".workflow"},
		{"cat <<EOF", ".workflow"},
		{"jq ", ".workflow/"},
	}
	for _, p := range patterns {
		if hit, idx := proximityHit(body, p.needle, p.near); hit {
			t.Fatalf("%s: forbidden pattern %q appears within %d chars of %q (near offset %d)",
				label, p.needle, proximityWindow, p.near, idx)
		}
	}
}

// forbiddingPrefixes mark a mention as an anti-pattern callout rather
// than a real authoring path. "Do NOT use `python3 -c`" must not trip
// the proximity check; "Run `python3 -c ...` to mutate .workflow" must.
var forbiddingPrefixes = []string{"Do NOT use", "Do not use", "do NOT use", "never use", "Never use", "forbid", "Forbid"}

func proximityHit(body, needle, near string) (bool, int) {
	idx := 0
	for {
		rel := strings.Index(body[idx:], needle)
		if rel < 0 {
			return false, -1
		}
		pos := idx + rel
		windowStart := pos - proximityWindow
		if windowStart < 0 {
			windowStart = 0
		}
		windowEnd := pos + len(needle) + proximityWindow
		if windowEnd > len(body) {
			windowEnd = len(body)
		}
		if strings.Contains(body[windowStart:windowEnd], near) &&
			!hasForbiddingPrefix(body[windowStart:pos]) {
			return true, pos
		}
		idx = pos + len(needle)
	}
}

func hasForbiddingPrefix(prefix string) bool {
	for _, p := range forbiddingPrefixes {
		if strings.Contains(prefix, p) {
			return true
		}
	}
	return false
}

func assertNoEmbeddedSkeleton(t *testing.T, label, body string) {
	t.Helper()
	// The embedded JSON skeleton, if present, would contain
	// `"feature": "<FEATURE_NAME>"` followed shortly by `"step": "`.
	const a = `"feature": "<FEATURE_NAME>"`
	if hit, _ := proximityHit(body, a, `"step": "`); hit {
		t.Fatalf("%s: embedded JSON skeleton still present (run `centinela evidence schema` instead)", label)
	}
}
