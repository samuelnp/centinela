package acceptance

import (
	"regexp"
	"strconv"
	"strings"
)

// cucumberSummary matches the run-summary line cucumber-js and godog both
// emit. It is anchored to the WHOLE line, so a summary-shaped string inside a
// test's own stdout or inside a failure message ("... got 2 scenarios (1
// skipped) from ...") can never match.
var cucumberSummary = regexp.MustCompile(`^(\d+) scenarios?(?: \(([^()]*)\))?[ \t]*\r?$`)

// cucumberCount matches one "N label" entry inside the parenthesised breakdown.
var cucumberCount = regexp.MustCompile(`^(\d+) ([a-z]+)$`)

// cucumberSummaryOf parses ONE line as a Gherkin run summary.
//
// It is line-scoped rather than whole-output because a Gherkin summary needs
// the same tier attribution a Go result does: under a whole-repo command the
// summary belongs to the package block that printed it, and counting a unit
// package's scenarios against the acceptance gate is exactly the over-block the
// Scope split exists to prevent.
func cucumberSummaryOf(line string) (Summary, bool) {
	m := cucumberSummary.FindStringSubmatch(line)
	if m == nil {
		return Summary{}, false
	}
	total, err := strconv.Atoi(m[1])
	if err != nil {
		return Summary{}, false
	}
	s := Summary{Shape: ShapeCucumber, Scenarios: total, SkipData: true, GherkinZero: total == 0}
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
