package workflow

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gatereport"
)

const unstamped = "### Report\n**Status:** SAFE\n\n```json centinela:verification\n" +
	`{"revision":"","treeDigest":"","commands":[{"argv":["centinela","validate"],"exitCode":0,"durationMs":84210}]}` +
	"\n```\n"

func TestStampVerificationFillsTheBlockAndKeepsCommands(t *testing.T) {
	run := gitStub("9f2c1ab", " M internal/x.go\n")
	seedGate(t, ValidateContractAdversarial, unstamped)
	snap, err := StampVerification("f", ".", run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Revision != "9f2c1ab" {
		t.Fatalf("revision = %q", snap.Revision)
	}
	data, _ := os.ReadFile(GatekeeperReportPath("f"))
	if !strings.Contains(string(data), `"durationMs":84210`) {
		t.Fatalf("commands array not preserved verbatim:\n%s", data)
	}
	v, err := gatereport.ParseVerification(string(data))
	if err != nil || v.Revision != snap.Revision || v.TreeDigest != snap.Digest {
		t.Fatalf("stamp not written: %+v %v", v, err)
	}
	// The stamp must make the report both admissible and fresh.
	if err := gatereport.Assess(string(data)); err != nil {
		t.Fatalf("stamped report inadmissible: %v", err)
	}
	if err := VerificationFresh("f", ".", run); err != nil {
		t.Fatalf("freshly stamped report is stale: %v", err)
	}
}

// The verifier writes its report, then stamps. That write must not stale it.
func TestStampVerificationSurvivesItsOwnWrite(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, unstamped)
	stamping := gitStub("9f2c1ab", "")
	if _, err := StampVerification("f", ".", stamping); err != nil {
		t.Fatal(err)
	}
	afterWrite := gitStub("9f2c1ab", " M .workflow/f-gatekeeper.md\n")
	if err := VerificationFresh("f", ".", afterWrite); err != nil {
		t.Fatalf("the stamp self-invalidated: %v", err)
	}
}

func TestStampVerificationMissingReport(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, unstamped)
	os.Remove(GatekeeperReportPath("f")) //nolint:errcheck
	_, err := StampVerification("f", ".", gitStub("a", ""))
	if err == nil || !strings.Contains(err.Error(), "artifact new") {
		t.Fatalf("want a not-found error naming the remedy, got %v", err)
	}
}

func TestStampVerificationMalformedBlock(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, "```json centinela:verification\n{nope\n```\n")
	if _, err := StampVerification("f", ".", gitStub("a", "")); err == nil {
		t.Fatal("want a malformed-block error")
	}
}

func TestStampVerificationSurfacesGitFailure(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, unstamped)
	boom := func(string, ...string) (string, error) { return "", os.ErrNotExist }
	if _, err := StampVerification("f", ".", boom); err == nil {
		t.Fatal("want a git failure")
	}
}
