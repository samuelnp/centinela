package workflow

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestVerificationStaleOnRevisionSkew(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, freshReport(t, gitStub("abc123", "")))
	err := VerificationFresh("f", ".", gitStub("def456", ""))
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("want stale error, got %v", err)
	}
	for _, sha := range []string{"abc123", "def456"} {
		if !strings.Contains(err.Error(), sha) {
			t.Fatalf("error must name %s: %v", sha, err)
		}
	}
	if !strings.Contains(err.Error(), "FRESH verifier") {
		t.Fatalf("error must demand a fresh verifier: %v", err)
	}
}

// The scenario a HEAD-only comparison would wrongly admit.
func TestVerificationStaleOnUncommittedSourceEdit(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, freshReport(t, gitStub("abc123", "")))
	err := VerificationFresh("f", ".", gitStub("abc123", " M internal/x.go\n"))
	if err == nil || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("want stale-tree error, got %v", err)
	}
}

func TestVerificationFreshDefersAbsentBlockToTheContentGate(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, "**Status:** SAFE\n")
	if err := VerificationFresh("f", ".", gitStub("abc123", "")); err != nil {
		t.Fatalf("an absent block is gatereport.Assess's message to give: %v", err)
	}
}

func TestVerificationFreshLegacyIsExempt(t *testing.T) {
	seedGate(t, "", freshReport(t, gitStub("abc123", "")))
	if err := VerificationFresh("f", ".", gitStub("zzz999", "")); err != nil {
		t.Fatalf("legacy workflows must be exempt: %v", err)
	}
}

func TestVerificationFreshSurfacesGitAndReadFailures(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, freshReport(t, gitStub("abc123", "")))
	boom := func(string, ...string) (string, error) { return "", errors.New("no git") }
	if err := VerificationFresh("f", ".", boom); err == nil || !strings.Contains(err.Error(), "no git") {
		t.Fatalf("want git failure surfaced, got %v", err)
	}
	seedGate(t, ValidateContractAdversarial, "")
	os.Remove(GatekeeperReportPath("f")) //nolint:errcheck
	if err := VerificationFresh("f", ".", gitStub("a", "")); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("want not-found error for an absent report, got %v", err)
	}
}
