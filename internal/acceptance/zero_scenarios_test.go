package acceptance

import (
	"strings"
	"testing"
)

// godogWrapped is the ordinary shape of godog driven from a Go wrapper test:
// a Gherkin summary AND go-test result lines in one report. Here the suite ran
// nothing — the wrapper passing is not evidence that any scenario executed.
const godogWrapped = `0 scenarios
0 steps
--- PASS: TestFeatures (0.00s)
PASS
ok  	gv/tests/acceptance	0.10s
`

// FINDING 1. A zero-scenario Gherkin run must stay failable no matter what
// unrelated passing signal shares the output. Combining by maximum alone would
// let the wrapper's Scenarios=1 rescue it.
func TestJudge_ZeroScenariosSurvivesAPassingGoSignal(t *testing.T) {
	v, detail := Judge("sh ./run.sh # tests/acceptance", godogWrapped, PolicyFail)
	if v != VerdictFail {
		t.Fatalf("a 0-scenario run must fail even beside a passing Go test, got %v (%q)", v, detail)
	}
	if !strings.Contains(detail, "no scenarios") {
		t.Fatalf("detail must state that no scenarios executed, got %q", detail)
	}
}

// The flag rides through Detect, so the fact survives the shape combination.
func TestDetect_GherkinZeroSurvivesTheCombination(t *testing.T) {
	s, ok := Detect(godogWrapped, ScopeAcceptance)
	if !ok || !s.GherkinZero {
		t.Fatalf("a 0-scenario Gherkin summary must be recorded, got %+v ok=%v", s, ok)
	}
	if s.Scenarios == 0 {
		t.Fatal("the Go half still contributed a scenario — this is why the flag is needed")
	}
}

// Bare zero-scenario output (no Go lines) is unchanged.
func TestJudge_BareZeroScenariosStillFails(t *testing.T) {
	if v, d := Judge("npx cucumber-js", "0 scenarios\n", PolicyFail); v != VerdictFail {
		t.Fatalf("a bare 0-scenario report must still fail, got %v (%q)", v, d)
	}
}

// The mirror: a real run that executed scenarios and skipped nothing is clean,
// and the zero flag is not set for it.
func TestJudge_WrappedNonZeroRunStaysGreen(t *testing.T) {
	out := strings.Replace(godogWrapped, "0 scenarios\n0 steps", "3 scenarios (3 passed)\n9 steps", 1)
	s, _ := Detect(out, ScopeAcceptance)
	if s.GherkinZero {
		t.Fatalf("a non-zero run must not be flagged, got %+v", s)
	}
	if v, d := Judge("sh ./run.sh # tests/acceptance", out, PolicyFail); v != VerdictPass {
		t.Fatalf("a clean wrapped godog run must pass, got %v (%q)", v, d)
	}
}

// The two "nothing ran" wordings are distinct: a Gherkin summary that reported
// zero, and a shape that yielded no results at all. Both must say "no scenarios".
func TestDescribe_BothNothingRanWordings(t *testing.T) {
	zero := Describe("c", Summary{Shape: ShapeCucumber, GherkinZero: true, Scenarios: 1}, ScopeAcceptance)
	if !strings.Contains(zero, "reported 0 scenarios") || !strings.Contains(zero, "no scenarios") {
		t.Fatalf("a zero-scenario Gherkin run must be named as such, got %q", zero)
	}
	empty := Describe("c", Summary{Shape: ShapeGoVerbose}, ScopeAcceptance)
	if !strings.Contains(empty, "executed no scenarios at all") {
		t.Fatalf("an empty report must be named as such, got %q", empty)
	}
}

// Under a whole-repo command the zero-scenario fact must itself be attributed:
// a UNIT package reporting 0 scenarios is not the acceptance suite's problem.
func TestJudge_ZeroScenariosInAnotherTierDoesNotFail(t *testing.T) {
	out := "0 scenarios\nPASS\nok  \tgv/unitpkg\t0.10s\n" +
		"--- PASS: TestAcceptOK (0.00s)\nPASS\nok  \tgv/tests/acceptance\t0.10s\n"
	if v, d := Judge("go test -v ./...", out, PolicyFail); v != VerdictPass {
		t.Fatalf("another tier's 0-scenario report must not fail the gate, got %v (%q)", v, d)
	}
}
