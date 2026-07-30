package acceptance

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
}

// Unexecuted counts the scenarios that were reported but did not assert.
func (s Summary) Unexecuted() int { return s.Skipped + s.Pending + s.Undefined }

// parsers are tried in order. Cucumber first: a godog run driven by `go test`
// emits BOTH a scenario summary and go's package lines, and the scenario
// summary is the more precise report. JSON before verbose before non-verbose,
// because each later shape's markers also appear inside the earlier ones.
var parsers = []func(string) (Summary, bool){
	parseCucumber,
	parseGoJSON,
	parseGoVerbose,
	parseGoNonVerbose,
}

// Detect returns the first matching runner summary. The false return means no
// supported shape matched — undetermined, which is never a pass verdict and
// never a failure verdict.
func Detect(output string) (Summary, bool) {
	for _, parse := range parsers {
		if s, ok := parse(output); ok {
			return s, true
		}
	}
	return Summary{}, false
}
