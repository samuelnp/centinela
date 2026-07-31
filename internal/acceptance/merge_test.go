package acceptance

import (
	"strings"
	"testing"
)

// mixedShapes is the verifier's exact repro for the first-match-wins false
// green: a clean cucumber summary printed alongside a go-level `--- SKIP:` —
// the natural shape of one acceptance package holding both godog scenarios and
// plain Go tests.
const mixedShapes = `3 scenarios (3 passed)
9 steps (9 passed)
=== RUN   TestGoLevelHidden
--- SKIP: TestGoLevelHidden (0.00s)
PASS
ok  	x/tests/acceptance	0.42s
`

// FINDING 2. Detect must not stop at the first matching parser: the Go skip is
// real evidence and discarding it renders a silent green.
func TestDetect_UnionsCucumberAndGoSkips(t *testing.T) {
	s, ok := Detect(mixedShapes, ScopeAcceptance)
	if !ok {
		t.Fatal("mixed output must be recognized")
	}
	if s.Skipped != 1 {
		t.Fatalf("the go-level skip must survive the cucumber match, got %+v", s)
	}
	if !strings.Contains(s.Shape, ShapeCucumber) || !strings.Contains(s.Shape, ShapeGoVerbose) {
		t.Fatalf("shape must name both contributing parsers, got %q", s.Shape)
	}
	// Counts must NOT be summed: the Go wrapper test re-describes the same run,
	// so "of 4 scenarios" would inflate the number the operator is told to act on.
	if s.Scenarios != 3 || s.Passed != 3 {
		t.Fatalf("counts must combine by maximum, not sum, got %+v", s)
	}
}

// Finding 4: no shape combination may inflate a total. One real skip described
// by -json, -v and a Gherkin summary at once is still one skip.
func TestMerge_DoesNotInflateOverlappingSignals(t *testing.T) {
	out := `1 scenarios (1 skipped)
--- SKIP: TestWrapper (0.00s)
{"Action":"skip","Package":"x/tests/acceptance","Test":"TestWrapper"}
ok  	x/tests/acceptance	0.10s
`
	s, ok := Detect(out, ScopeAcceptance)
	if !ok {
		t.Fatal("output must be recognized")
	}
	if s.Skipped != 1 {
		t.Fatalf("one real skip must be reported once, got %d (%+v)", s.Skipped, s)
	}
	if s.Scenarios != 1 {
		t.Fatalf("one real run must not be counted per shape, got %d (%+v)", s.Scenarios, s)
	}
}

func TestJudge_MixedShapesIsNotASilentGreen(t *testing.T) {
	v, detail := Judge("./run.sh # tests/acceptance", mixedShapes, PolicyFail)
	if v != VerdictFail {
		t.Fatalf("a go skip beside a clean cucumber summary must fail, got %v (%q)", v, detail)
	}
	if !strings.Contains(detail, "1 skipped") {
		t.Fatalf("detail must name the hidden skip, got %q", detail)
	}
}

// The union must not invent skips: two clean shapes stay clean.
func TestDetect_UnionOfCleanShapesStaysClean(t *testing.T) {
	out := "3 scenarios (3 passed)\n--- PASS: TestA (0.00s)\nok  \tx/tests/acceptance\t0.1s\n"
	s, ok := Detect(out, ScopeAcceptance)
	if !ok || s.Unexecuted() != 0 {
		t.Fatalf("clean shapes must union to a clean summary, got %+v ok=%v", s, ok)
	}
	if v, d := Judge("npx cucumber-js", out, PolicyFail); v != VerdictPass || d != "" {
		t.Fatalf("a clean union must pass silently, got %v (%q)", v, d)
	}
}

// The non-verbose fallback only applies when NO skip-data shape matched, so a
// `-v` run never degrades into the "carries no skip data" note.
func TestDetect_NonVerboseIsOnlyAFallback(t *testing.T) {
	s, ok := Detect(mixedShapes, ScopeAcceptance)
	if !ok || !s.SkipData {
		t.Fatalf("a shape carrying skip data must win over the fallback, got %+v", s)
	}
	plain, ok := Detect("ok  \tx/a\t0.1s\n", ScopeAcceptance)
	if !ok || plain.SkipData || plain.Shape != ShapeGoNonVerbose {
		t.Fatalf("plain output must still reach the fallback, got %+v", plain)
	}
}
