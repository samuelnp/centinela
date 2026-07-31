package acceptance

import (
	"encoding/json"
	"strings"
)

// goJSONEvent is the subset of `go test -json` records this parser reads.
// Test is empty on PACKAGE-level events: a package-level "skip" is the
// "[no test files]" case, which is not a skipped scenario. Package is what
// makes tier attribution possible under ScopeMixed.
type goJSONEvent struct {
	Action  string `json:"Action"`
	Test    string `json:"Test"`
	Package string `json:"Package"`
}

// parseGoJSON counts test-level results, keeping only those in the tier under
// analysis. Events arrive interleaved when packages run in parallel; because
// the format is one self-contained JSON object per line, interleaving cannot
// corrupt the counts and every event carries its own package.
func parseGoJSON(output string, scope Scope) (Summary, bool) {
	s := Summary{Shape: ShapeGoJSON, SkipData: true, Attributed: scope == ScopeMixed}
	seen := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var e goJSONEvent
		if json.Unmarshal([]byte(line), &e) != nil || e.Action == "" {
			continue
		}
		seen = true
		if e.Test == "" || !scope.counts(e.Package) {
			continue
		}
		addGoResult(&s, strings.ToUpper(e.Action))
	}
	return s, seen
}

// addGoResult folds one per-test result into a summary. A failing test is an
// executed scenario, not a skipped one; anything else (run, output, start) is
// not a scenario at all.
func addGoResult(s *Summary, result string) {
	switch result {
	case "PASS":
		s.Passed++
		s.Scenarios++
	case "SKIP":
		s.Skipped++
		s.Scenarios++
	case "FAIL":
		s.Scenarios++
	}
}
