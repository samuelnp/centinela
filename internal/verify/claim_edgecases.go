package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/samuelnp/centinela/internal/evidence"
)

const claimEdgeCases = "edge-cases-to-tests"

var (
	testNameRe = regexp.MustCompile(`func\s+(Test\w+)\s*\(`)
	nonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)
)

// checkEdgeCases cross-checks each edgeCases entry against the test function
// names in the feature's test files. WARN-only (heuristic string matching): an
// unmatched edge case is reported but never hard-blocks. Empty list → SKIP.
func checkEdgeCases(root, role string, ev *evidence.RoleEvidence) Check {
	c := Check{Claim: claimEdgeCases, Role: role}
	if len(ev.EdgeCases) == 0 {
		c.Status = StatusSkip
		c.Detail = "no edgeCases claimed"
		return c
	}
	haystack := collectTestNames(root)
	var unmatched []string
	for _, ec := range ev.EdgeCases {
		if !edgeCaseMatches(ec, haystack) {
			unmatched = append(unmatched, ec)
		}
	}
	if len(unmatched) > 0 {
		c.Status = StatusWarn
		c.Detail = "no matching test found for: " + strings.Join(unmatched, "; ")
		return c
	}
	c.Status = StatusPass
	c.Detail = fmt.Sprintf("all %d edge case(s) map to a test name", len(ev.EdgeCases))
	return c
}

// edgeCaseMatches reports whether any significant word (>= 4 chars) of the
// edge-case phrase appears in the normalized concatenation of test names.
func edgeCaseMatches(edgeCase, haystack string) bool {
	for _, word := range strings.Fields(strings.ToLower(edgeCase)) {
		w := nonAlnumRe.ReplaceAllString(word, "")
		if len(w) >= 4 && strings.Contains(haystack, w) {
			return true
		}
	}
	return false
}

// collectTestNames walks root for *_test.go files and returns every Test
// function name normalized to a single lowercase-alphanumeric blob.
func collectTestNames(root string) string {
	var b strings.Builder
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		for _, m := range testNameRe.FindAllStringSubmatch(string(data), -1) {
			b.WriteString(nonAlnumRe.ReplaceAllString(strings.ToLower(m[1]), ""))
			b.WriteByte(' ')
		}
		return nil
	})
	return b.String()
}
