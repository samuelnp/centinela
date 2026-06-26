package delivery

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/verify"
)

const secBrief = "## Problem\n\nbig problem\n\n## Who / Why\n\nfor users\n\n## Acceptance Summary\n\nit works\n"

func TestSummarySectionBriefAndPlanFallback(t *testing.T) {
	if got := summarySection(Evidence{Brief: secBrief}); !strings.Contains(got, "## Summary") || !strings.Contains(got, "big problem") {
		t.Fatalf("brief summary: %q", got)
	}
	plan := "## Problem & Goal\n\nplan goal\n"
	if got := summarySection(Evidence{Plan: plan}); !strings.Contains(got, "plan goal") {
		t.Fatalf("plan fallback: %q", got)
	}
	if summarySection(Evidence{}) != "" {
		t.Fatal("summary omitted when absent")
	}
}

func TestWhatWhySection(t *testing.T) {
	e := Evidence{Brief: secBrief, Plan: "## Proposed Architecture\n\narch text\n"}
	got := whatWhySection(e)
	if !strings.Contains(got, "for users") || !strings.Contains(got, "arch text") {
		t.Fatalf("what/why both halves: %q", got)
	}
	if whatWhySection(Evidence{}) != "" {
		t.Fatal("what/why omitted when absent")
	}
}

func TestAcceptanceSection(t *testing.T) {
	if got := acceptanceSection(Evidence{Brief: secBrief, SpecPath: "specs/a.feature"}); !strings.Contains(got, "it works") || !strings.Contains(got, "specs/a.feature") {
		t.Fatalf("acceptance both: %q", got)
	}
	if got := acceptanceSection(Evidence{SpecPath: "specs/a.feature"}); !strings.Contains(got, "specs/a.feature") {
		t.Fatalf("pointer only: %q", got)
	}
	if acceptanceSection(Evidence{}) != "" {
		t.Fatal("acceptance omitted when absent")
	}
}

func TestGateStatusSection(t *testing.T) {
	if got := gateStatusSection(Evidence{GatekeeperReport: "**Status:** SAFE"}); !strings.Contains(got, "SAFE") {
		t.Fatalf("verdict: %q", got)
	}
	v := &verify.VerificationResult{Checks: []verify.Check{{Status: verify.StatusPass}, {Status: verify.StatusFail}}}
	if got := gateStatusSection(Evidence{Verification: v}); !strings.Contains(got, "1 pass, 1 fail") {
		t.Fatalf("tally: %q", got)
	}
	if gateStatusSection(Evidence{}) != "" {
		t.Fatal("gate status omitted when neither present")
	}
}

func TestProvenanceFooterAlways(t *testing.T) {
	if got := provenanceFooter(Evidence{Feature: "alpha"}); !strings.Contains(got, "alpha") {
		t.Fatalf("provenance: %q", got)
	}
}
