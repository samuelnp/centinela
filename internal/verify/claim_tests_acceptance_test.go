package verify

import (
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

const skipReport = "3 scenarios (1 skipped, 2 passed)\n"

// AC6: an acceptance-classified command that exits 0 while reporting a skip is
// a claim FAIL, distinct from PASS.
func TestCheckTestsPass_AcceptanceSkipIsAClaimFailure(t *testing.T) {
	cfg := cfgWithCmds("npx cucumber-js")
	cfg.Validate.AcceptanceSkipPolicy = config.AcceptanceSkipFail
	deps := Deps{Runner: &fakeRunner{def: RunOutcome{ExitCode: 0, Output: skipReport}}}
	got := checkTestsPass(cfg, deps, "qa-senior", time.Second)
	if got.Status != StatusFail {
		t.Fatalf("status = %q, want FAIL (detail %q)", got.Status, got.Detail)
	}
	if got.Status == StatusPass {
		t.Fatal("the skip verdict must be distinct from PASS")
	}
	if !strings.Contains(got.Detail, "1 skipped") {
		t.Fatalf("detail must name the skipped count, got %q", got.Detail)
	}
}

// R8: the reused single-run outcome is labelled "validate.commands", not a real
// command, so it is classified over the CONFIGURED SET. Getting this wrong
// silently disarms the rule on the exact path `complete` uses.
func TestCheckTestsPass_PriorRunAppliesTheAcceptanceRule(t *testing.T) {
	cfg := cfgWithCmds("go vet ./...", "npx cucumber-js")
	prior := &RunOutcome{ExitCode: 0, Output: skipReport}
	deps := Deps{Runner: &fakeRunner{}, PriorTestRun: prior}
	if got := checkTestsPass(cfg, deps, "qa", time.Second); got.Status != StatusFail {
		t.Fatalf("reused outcome must be analysed for acceptance skips, got %q / %q",
			got.Status, got.Detail)
	}
}

// AC5 on the verify path: with no acceptance-classified command configured, the
// same reused output is untouched.
func TestCheckTestsPass_PriorRunWithNoAcceptanceCommandPasses(t *testing.T) {
	cfg := cfgWithCmds("go vet ./...", "./scripts/check-fmt.sh")
	deps := Deps{Runner: &fakeRunner{}, PriorTestRun: &RunOutcome{ExitCode: 0, Output: skipReport}}
	if got := checkTestsPass(cfg, deps, "qa", time.Second); got.Status != StatusPass {
		t.Fatalf("non-acceptance commands must not be skip-analysed, got %q / %q",
			got.Status, got.Detail)
	}
}

// The verifier does not invent a failure it cannot prove.
func TestCheckTestsPass_UnparseableStaysPassAndNamesTheLimitation(t *testing.T) {
	cfg := cfgWithCmds("npx cucumber-js")
	deps := Deps{Runner: &fakeRunner{def: RunOutcome{ExitCode: 0, Output: "Ran 12 examples\n"}}}
	got := checkTestsPass(cfg, deps, "qa", time.Second)
	if got.Status != StatusPass {
		t.Fatalf("an unprovable skip must stay PASS, got %q", got.Status)
	}
	if !strings.Contains(got.Detail, "could not be parsed") {
		t.Fatalf("detail must name the limitation, got %q", got.Detail)
	}
}

// The warn policy surfaces the counts on this path without failing the claim.
func TestCheckTestsPass_WarnPolicyDoesNotFailTheClaim(t *testing.T) {
	cfg := cfgWithCmds("npx cucumber-js")
	cfg.Validate.AcceptanceSkipPolicy = config.AcceptanceSkipWarn
	deps := Deps{Runner: &fakeRunner{def: RunOutcome{ExitCode: 0, Output: skipReport}}}
	got := checkTestsPass(cfg, deps, "qa", time.Second)
	if got.Status != StatusPass || !strings.Contains(got.Detail, "1 skipped") {
		t.Fatalf("warn policy must pass while naming the counts, got %q / %q",
			got.Status, got.Detail)
	}
}
