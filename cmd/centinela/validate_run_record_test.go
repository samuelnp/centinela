package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// FINDING 3, producer side. The record must carry the REAL captured output and
// the judgement the per-command analysis produced, so `complete` reuses an
// analysis that genuinely ran instead of re-parsing fixed prose.
func TestValidateRunRecord_CarriesTranscriptAndJudgement(t *testing.T) {
	var r validateRunRecord
	r.record("npx cucumber-js", "3 scenarios (3 passed)", true, gates.Pass, "", config.AcceptanceSkipFail)
	r.record("./scripts/check-fmt.sh", "", true, gates.Pass, "", config.AcceptanceSkipFail)
	j := r.judgement()
	if !j.Analysed || j.Failed {
		t.Fatalf("a clean acceptance command is analysed and not failed, got %+v", j)
	}
	if !strings.Contains(r.transcript(), "3 scenarios") {
		t.Fatalf("the real output must be carried, got %q", r.transcript())
	}
}

func TestValidateRunRecord_SkipFailureIsRecorded(t *testing.T) {
	var r validateRunRecord
	r.record("npx cucumber-js", "3 scenarios (1 skipped, 2 passed)", true, gates.Fail, "reported 1 skipped", config.AcceptanceSkipFail)
	j := r.judgement()
	if !j.Analysed || !j.Failed || !strings.Contains(j.Detail, "1 skipped") {
		t.Fatalf("a skip-driven failure must be recorded with its detail, got %+v", j)
	}
}

// An EXIT-code failure is not a skip verdict, and must not be recorded as one.
func TestValidateRunRecord_ExitFailureIsNotASkipFailure(t *testing.T) {
	var r validateRunRecord
	r.record("npx cucumber-js", "3 scenarios (1 skipped)", false, gates.Fail, "", config.AcceptanceSkipFail)
	if r.judgement().Failed {
		t.Fatal("a non-zero exit must never be recorded as an acceptance skip failure")
	}
}

// No acceptance-classified command means no analysis was performed — the
// judgement must say so rather than imply a clean result.
func TestValidateRunRecord_NoAcceptanceCommandIsNotAnalysed(t *testing.T) {
	var r validateRunRecord
	r.record("go vet ./...", "", true, gates.Pass, "", config.AcceptanceSkipFail)
	if j := r.judgement(); j.Analysed {
		t.Fatalf("no acceptance command means no analysis, got %+v", j)
	}
}

// Under the off policy nothing is parsed, so Analysed must stay FALSE — that
// field's whole job is keeping "ran and found nothing" distinct from "never
// ran", and the reuse message must name the real reason.
func TestValidateRunRecord_OffPolicyIsNotAnAnalysis(t *testing.T) {
	var r validateRunRecord
	r.record("npx cucumber-js", "3 scenarios (1 skipped, 2 passed)", true, gates.Pass, "",
		config.AcceptanceSkipOff)
	j := r.judgement()
	if j.Analysed || j.Failed {
		t.Fatalf("the off policy performs no analysis, got %+v", j)
	}
	if !strings.Contains(j.Detail, "disabled by") {
		t.Fatalf("detail must name the policy as the reason, got %q", j.Detail)
	}
}

// End-to-end: running the commands populates the record the reused outcome reads.
func TestCompletedValidationOutcome_ReflectsTheRunItStandsFor(t *testing.T) {
	cfg := &config.Config{}
	cfg.Validate.Commands = []string{`printf '3 scenarios (1 skipped, 2 passed)\n' # acceptance test`}
	cfg.Validate.AcceptanceSkipPolicy = config.AcceptanceSkipFail
	captureStdout(t, func() { runValidateCommands(cfg) })
	outcome := completedValidationOutcome()
	if outcome.AcceptanceJudged == nil || !outcome.AcceptanceJudged.Failed {
		t.Fatalf("the reused outcome must carry the real judgement, got %+v", outcome.AcceptanceJudged)
	}
	if !strings.Contains(outcome.Output, "verified tree") ||
		!strings.Contains(outcome.Output, "1 skipped") {
		t.Fatalf("Output must keep the provenance header AND the real transcript, got %q", outcome.Output)
	}
}
