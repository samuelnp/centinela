package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/treestate"
)

// gitStub answers rev-parse with head and reports a working tree that is
// clean apart from the given porcelain status. A revision range reports a
// SOURCE file, so a moved HEAD means real product churn — the roadmap-state
// exemption (D5) is opted into explicitly by gitStubRange.
func gitStub(head, status string) treestate.Runner {
	return gitStubRange(head, status, "internal/x.go\n")
}

// gitStubRange additionally controls what `diff --name-only <rev>..HEAD`
// answers, which is what decides whether a moved HEAD is exempt.
func gitStubRange(head, status, rangePaths string) treestate.Runner {
	return func(_ string, args ...string) (string, error) {
		switch {
		case args[0] == "rev-parse":
			return head + "\n", nil
		case args[0] == "status":
			return status, nil
		case args[0] == "diff" && len(args) > 1 && args[1] == "--name-only":
			return rangePaths, nil
		default:
			return "", nil
		}
	}
}

// freshReport stamps a report against the tree gitStub describes, so the test
// asserts the comparison rather than re-deriving the digest by hand.
func freshReport(t *testing.T, run treestate.Runner) string {
	t.Helper()
	snap, err := treestate.Stamp(".", run)
	if err != nil {
		t.Fatal(err)
	}
	return "**Status:** SAFE\n\n```json centinela:verification\n{\"revision\":\"" +
		snap.Revision + "\",\"treeDigest\":\"" + snap.Digest +
		"\",\"commands\":[{\"argv\":[\"centinela\",\"validate\"],\"exitCode\":0}]}\n```\n"
}

func TestVerificationFreshMatchingStamp(t *testing.T) {
	run := gitStub("abc123", "")
	seedGate(t, ValidateContractAdversarial, freshReport(t, run))
	if err := VerificationFresh("f", ".", run); err != nil {
		t.Fatalf("matching stamp must be fresh: %v", err)
	}
}

// D3a, the direction that would deadlock a real run if it regressed.
func TestVerificationFreshSurvivesWorkflowOnlyChurn(t *testing.T) {
	stamped := gitStub("abc123", "")
	seedGate(t, ValidateContractAdversarial, freshReport(t, stamped))
	churned := gitStub("abc123", " M .workflow/f-gatekeeper.md\n?? .workflow/f-gatekeeper.json\n")
	if err := VerificationFresh("f", ".", churned); err != nil {
		t.Fatalf(".workflow/-only churn must not stale the verification: %v", err)
	}
}
