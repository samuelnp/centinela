package acceptance

import "strings"

// Shape names identify which parser matched a report.
const (
	ShapeCucumber     = "cucumber/godog summary"
	ShapeGoJSON       = "go test -json"
	ShapeGoVerbose    = "go test -v"
	ShapeGoNonVerbose = "go test (non-verbose)"
)

// Summary is a parsed runner report.
//
// SkipData is the honest half of this type. A shape can be RECOGNIZED and still
// carry no skip information: plain non-verbose `go test` prints one ok/FAIL line
// per package and nothing about skipped tests. Reporting that as "could not be
// parsed" would put a permanent ⚠ on every run of a shape that can never carry
// the data, and a standing ⚠ trains operators to ignore ⚠. So a recognized
// shape without skip data is a quiet pass with a note, and the warning is
// reserved for output that matches no known runner at all.
type Summary struct {
	Shape     string
	Scenarios int
	Passed    int
	Skipped   int
	Pending   int
	Undefined int
	SkipData  bool
	// GherkinZero records that an ATTRIBUTED Gherkin summary reported zero
	// scenarios. It is tracked per shape rather than derived from the merged
	// Scenarios total, because a suite that ran nothing must stay failable no
	// matter what unrelated passing signal shares the same output — a Go
	// wrapper test's `--- PASS:` is not evidence that any scenario executed.
	GherkinZero bool
	// Attributed records that tier filtering was applied to these counts, so
	// the message can say so without asserting an attribution that never ran.
	Attributed bool
}

// Unexecuted counts the scenarios that were reported but did not assert.
func (s Summary) Unexecuted() int { return s.Skipped + s.Pending + s.Undefined }

// Detect combines every recognized skip-data shape present in the output, and
// only falls back to the skip-data-free shapes when none matched. A false
// return means no supported shape matched at all — undetermined, which is
// never a pass verdict and never a failure verdict.
//
// Stopping at the first match is what let a run printing a clean cucumber
// summary alongside go-level `--- SKIP:` lines render green with the Go skips
// silently discarded, so every shape contributes; but the contributions are
// combined by maximum, not sum, because they routinely re-describe one run.
func Detect(output string, scope Scope) (Summary, bool) {
	var matched []Summary
	if s, ok := parseGoJSON(output, scope); ok {
		matched = append(matched, s)
	}
	gherkin, gotest, sawGherkin, sawGo := scanTextReport(output, scope)
	if sawGherkin {
		matched = append(matched, gherkin)
	}
	if sawGo {
		matched = append(matched, gotest)
	}
	if len(matched) > 0 {
		return merge(matched), true
	}
	return parseGoNonVerbose(output, scope)
}

// merge combines the shapes present in one report without inflating any count.
func merge(list []Summary) Summary {
	out := Summary{SkipData: true}
	names := make([]string, 0, len(list))
	for _, s := range list {
		names = append(names, s.Shape)
		out.atLeast(s)
	}
	out.Shape = strings.Join(names, " + ")
	return out
}
