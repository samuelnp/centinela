package verify

import (
	"strings"
	"testing"
	"time"
)

// FINDING 3. The reuse path must act on the analysis the producer actually
// performed, not re-parse a joined transcript. A producer-reported skip failure
// is a claim FAIL.
func TestCheckTestsPass_InheritsAFailedProducerJudgement(t *testing.T) {
	cfg := cfgWithCmds("go test ./... -coverprofile=c.out")
	prior := &RunOutcome{
		ExitCode:         0,
		Output:           "validate.commands executed by this complete gate at the verified tree",
		AcceptanceJudged: &AcceptanceJudgement{Analysed: true, Failed: true, Detail: "reported 1 skipped"},
	}
	got := checkTestsPass(cfg, Deps{Runner: &fakeRunner{}, PriorTestRun: prior}, "qa", time.Second)
	if got.Status != StatusFail {
		t.Fatalf("a producer-reported skip failure must be a claim FAIL, got %q / %q", got.Status, got.Detail)
	}
	if !strings.Contains(got.Detail, "1 skipped") {
		t.Fatalf("detail must carry the producer's counts, got %q", got.Detail)
	}
}

// A clean judgement says the analysis RAN — it must never read as "could not be
// parsed", which is what the fixed prose Output used to produce on every run.
func TestCheckTestsPass_InheritedCleanJudgementSaysAnalysisRan(t *testing.T) {
	cfg := cfgWithCmds("go test ./... -coverprofile=c.out")
	prior := &RunOutcome{
		ExitCode:         0,
		Output:           "validate.commands executed by this complete gate at the verified tree",
		AcceptanceJudged: &AcceptanceJudgement{Analysed: true, Detail: "no skipped scenarios detected"},
	}
	got := checkTestsPass(cfg, Deps{Runner: &fakeRunner{}, PriorTestRun: prior}, "qa", time.Second)
	if got.Status != StatusPass {
		t.Fatalf("a clean judgement is a PASS, got %q", got.Status)
	}
	if !strings.Contains(got.Detail, "already performed") {
		t.Fatalf("detail must say the analysis ran, got %q", got.Detail)
	}
	if strings.Contains(got.Detail, "could not be parsed") {
		t.Fatalf("a performed analysis must not be reported as unparseable: %q", got.Detail)
	}
}

// When no analysis ran, the message says so plainly instead of blaming the
// report shape.
func TestCheckTestsPass_InheritedUnanalysedSaysNotPerformed(t *testing.T) {
	cfg := cfgWithCmds("go vet ./...")
	prior := &RunOutcome{ExitCode: 0, AcceptanceJudged: &AcceptanceJudgement{}}
	got := checkTestsPass(cfg, Deps{Runner: &fakeRunner{}, PriorTestRun: prior}, "qa", time.Second)
	if got.Status != StatusPass || !strings.Contains(got.Detail, "was not performed") {
		t.Fatalf("an unanalysed reuse must say so, got %q / %q", got.Status, got.Detail)
	}
}

// The off policy is "never ran", not "ran and found nothing", and the reuse
// message must name the policy rather than imply a clean analysis.
func TestCheckTestsPass_InheritedDisabledPolicySaysWhy(t *testing.T) {
	cfg := cfgWithCmds("npx cucumber-js")
	prior := &RunOutcome{ExitCode: 0, AcceptanceJudged: &AcceptanceJudgement{
		Detail: "skip detection disabled by [validate] acceptance_skip_policy = off"}}
	got := checkTestsPass(cfg, Deps{Runner: &fakeRunner{}, PriorTestRun: prior}, "qa", time.Second)
	if got.Status != StatusPass {
		t.Fatalf("the off policy is not a failure, got %q", got.Status)
	}
	if !strings.Contains(got.Detail, "was not performed") ||
		!strings.Contains(got.Detail, "acceptance_skip_policy = off") {
		t.Fatalf("detail must say it did not run AND why, got %q", got.Detail)
	}
	if strings.Contains(got.Detail, "already performed") {
		t.Fatalf("a disabled analysis must never claim it ran: %q", got.Detail)
	}
}
