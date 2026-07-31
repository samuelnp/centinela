package verify

import (
	"testing"
	"time"
)

// Without a producer judgement the fallback still applies. It resolves the
// scope from the configured set: a bare Gherkin transcript from an
// acceptance-scoped command counts, while a whole-repo command makes the whole
// joined transcript mixed so another tier's skips cannot over-block.
func TestCheckTestsPass_PriorRunFallbackScopeFromCommandSet(t *testing.T) {
	gherkin := "3 scenarios (1 skipped, 2 passed)\n"
	scoped := cfgWithCmds("go vet ./...", "npx cucumber-js")
	deps := Deps{Runner: &fakeRunner{}, PriorTestRun: &RunOutcome{ExitCode: 0, Output: gherkin}}
	if got := checkTestsPass(scoped, deps, "qa", time.Second); got.Status != StatusFail {
		t.Fatalf("an all-acceptance-scoped set must still see a bare Gherkin skip, got %q", got.Status)
	}
	mixed := cfgWithCmds("go test ./...", "npx cucumber-js")
	if got := checkTestsPass(mixed, deps, "qa", time.Second); got.Status != StatusPass {
		t.Fatalf("one whole-repo command makes the joined transcript unattributable, got %q / %q",
			got.Status, got.Detail)
	}
}

func TestCheckTestsPass_PriorRunFallbackUsesMixedScope(t *testing.T) {
	unit := "--- SKIP: TestUnitSkip (0.00s)\nok  \tgv/unitpkg\t0.1s\n"
	cfg := cfgWithCmds("go test ./...")
	deps := Deps{Runner: &fakeRunner{}, PriorTestRun: &RunOutcome{ExitCode: 0, Output: unit}}
	if got := checkTestsPass(cfg, deps, "qa", time.Second); got.Status != StatusPass {
		t.Fatalf("a unit-tier skip must not fail the reuse fallback, got %q / %q", got.Status, got.Detail)
	}
	accept := "--- SKIP: TestA (0.00s)\nok  \tgv/tests/acceptance\t0.1s\n"
	deps.PriorTestRun = &RunOutcome{ExitCode: 0, Output: accept}
	if got := checkTestsPass(cfg, deps, "qa", time.Second); got.Status != StatusFail {
		t.Fatalf("an acceptance-tier skip must still fail the fallback, got %q", got.Status)
	}
}
