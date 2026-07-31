package acceptance

import (
	"strings"
	"testing"
)

// unitGherkin: a whole-repo `go test -v ./...` where the UNIT package prints a
// scenario summary with a skip and the acceptance package is completely clean.
const unitGherkin = `2 scenarios (1 skipped, 1 passed)
PASS
ok  	gv/unitpkg	0.10s
--- PASS: TestAcceptOK (0.00s)
PASS
ok  	gv/tests/acceptance	0.10s
`

// FINDING 2, over-block half. A Gherkin summary is attributed to the package
// block that printed it, exactly like a Go result — otherwise a whole-repo
// command fails on another tier's Gherkin run while the acceptance tier is
// clean, and AC5's tier-level "never" is false.
func TestJudge_GherkinSummaryInAnotherTierDoesNotFail(t *testing.T) {
	v, detail := Judge("go test -v ./...", unitGherkin, PolicyFail)
	if v != VerdictPass {
		t.Fatalf("a unit package's Gherkin skip must not fail the gate, got %v (%q)", v, detail)
	}
}

// The mirror: the same summary printed by the ACCEPTANCE package still fails.
func TestJudge_GherkinSummaryInTheAcceptanceTierStillFails(t *testing.T) {
	out := strings.Replace(unitGherkin, "gv/unitpkg", "gv/tests/acceptance", 1)
	v, detail := Judge("go test -v ./...", out, PolicyFail)
	if v != VerdictFail {
		t.Fatalf("the acceptance package's Gherkin skip must fail, got %v (%q)", v, detail)
	}
	if !strings.Contains(detail, "1 skipped") {
		t.Fatalf("detail must name the count, got %q", detail)
	}
}

// FINDING 2, message half. The attribution clause may only appear when tier
// filtering actually ran — an acceptance-scoped command filters nothing.
func TestDescribe_AttributionClauseOnlyWhenAttributionHappened(t *testing.T) {
	mixed, ok := Detect(strings.Replace(unitGherkin, "gv/unitpkg", "gv/tests/acceptance", 1), ScopeMixed)
	if !ok || !mixed.Attributed {
		t.Fatalf("a whole-repo parse must record that it attributed, got %+v", mixed)
	}
	if !strings.Contains(Describe("c", mixed, ScopeMixed), "attributed to "+AcceptancePath) {
		t.Fatal("a filtered run must say so")
	}
	scoped, ok := Detect("2 scenarios (1 skipped, 1 passed)\n", ScopeAcceptance)
	if !ok || scoped.Attributed {
		t.Fatalf("an acceptance-scoped parse filters nothing, got %+v", scoped)
	}
	if strings.Contains(Describe("c", scoped, ScopeAcceptance), "attributed") {
		t.Fatalf("an unfiltered run must not claim attribution: %q", Describe("c", scoped, ScopeAcceptance))
	}
}

// A bare Gherkin runner has no package blocks at all; under ScopeAcceptance its
// trailing block is still trusted, and it claims no attribution.
func TestDetect_BareGherkinRunNeedsNoTerminator(t *testing.T) {
	s, ok := Detect("2 scenarios (1 skipped, 1 passed)\n8 steps (8 passed)\n", ScopeAcceptance)
	if !ok || s.Skipped != 1 || s.Attributed {
		t.Fatalf("a bare cucumber run must parse unfiltered, got %+v ok=%v", s, ok)
	}
}
