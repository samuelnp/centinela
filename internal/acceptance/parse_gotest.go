package acceptance

import (
	"encoding/json"
	"regexp"
	"strings"
)

// goJSONEvent is the subset of `go test -json` records this parser reads.
// Test is empty on PACKAGE-level events: a package-level "skip" is the
// "[no test files]" case, which is not a skipped scenario.
type goJSONEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
}

// parseGoJSON counts test-level results. Events arrive interleaved when
// packages run in parallel; because the format is one JSON object per line,
// interleaving cannot corrupt the counts.
func parseGoJSON(output string) (Summary, bool) {
	s := Summary{Shape: ShapeGoJSON, SkipData: true}
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
		if e.Test == "" {
			continue
		}
		countGoAction(&s, e.Action)
	}
	return s, seen
}

func countGoAction(s *Summary, action string) {
	switch action {
	case "pass":
		s.Passed++
		s.Scenarios++
	case "skip":
		s.Skipped++
		s.Scenarios++
	case "fail":
		s.Scenarios++
	}
}

// goVerboseResult matches `go test -v` per-test result lines, including the
// indented form subtests use.
var goVerboseResult = regexp.MustCompile(`(?m)^[ \t]*--- (PASS|FAIL|SKIP): `)

func parseGoVerbose(output string) (Summary, bool) {
	matches := goVerboseResult.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return Summary{}, false
	}
	s := Summary{Shape: ShapeGoVerbose, SkipData: true}
	for _, m := range matches {
		s.Scenarios++
		switch m[1] {
		case "PASS":
			s.Passed++
		case "SKIP":
			s.Skipped++
		}
	}
	return s, true
}

// goPackageLine matches the per-package result lines plain `go test` prints.
var goPackageLine = regexp.MustCompile(`(?m)^(ok|FAIL|\?)[ \t]+\S+[ \t]`)

// parseGoNonVerbose RECOGNIZES plain `go test` output and reports SkipData
// false: the shape is known, but it structurally carries no skip information.
// That is a quiet "no skip data", not a warning — see Summary.SkipData.
func parseGoNonVerbose(output string) (Summary, bool) {
	if !goPackageLine.MatchString(output) {
		return Summary{}, false
	}
	return Summary{Shape: ShapeGoNonVerbose}, true
}
