package acceptance

import (
	"regexp"
	"strconv"
	"strings"
)

// cucumberSummary matches the run-summary line cucumber-js and godog both
// emit. It is ANCHORED at the start of a line and consumes the whole line, so
// a summary-shaped string inside a test's own stdout or inside a failure
// message ("... got 2 scenarios (1 skipped) from ...") can never match.
var cucumberSummary = regexp.MustCompile(`(?m)^(\d+) scenarios?(?: \(([^()]*)\))?[ \t]*\r?$`)

// cucumberCount matches one "N label" entry inside the parenthesised breakdown.
var cucumberCount = regexp.MustCompile(`^(\d+) ([a-z]+)$`)

// parseCucumber reads the LAST summary line in the output: a run that prints
// per-feature summaries ends with the aggregate one.
func parseCucumber(output string) (Summary, bool) {
	matches := cucumberSummary.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return Summary{}, false
	}
	m := matches[len(matches)-1]
	total, err := strconv.Atoi(m[1])
	if err != nil {
		return Summary{}, false
	}
	s := Summary{Shape: ShapeCucumber, Scenarios: total, SkipData: true}
	applyCucumberCounts(&s, m[2])
	return s, true
}

// applyCucumberCounts folds the "1 skipped, 2 passed" breakdown into the
// summary. An unknown label is ignored rather than guessed at.
func applyCucumberCounts(s *Summary, breakdown string) {
	for _, part := range strings.Split(breakdown, ",") {
		m := cucumberCount.FindStringSubmatch(strings.TrimSpace(part))
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		switch m[2] {
		case "passed":
			s.Passed = n
		case "skipped":
			s.Skipped = n
		case "pending":
			s.Pending = n
		case "undefined":
			s.Undefined = n
		}
	}
}
