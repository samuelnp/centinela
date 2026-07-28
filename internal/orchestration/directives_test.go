package orchestration

import (
	"strings"
	"testing"
)

func TestDelegationContractValidateNamesTheContract(t *testing.T) {
	got := DelegationContract("validate")
	for _, want := range []string{"ADVERSARIAL VERIFIER", "FRESH context", "file paths", "do NOT summarize", "artifact stamp"} {
		if !strings.Contains(got, want) {
			t.Fatalf("validate contract missing %q: %s", want, got)
		}
	}
}

func TestDelegationContractEmptyForOtherSteps(t *testing.T) {
	for _, step := range []string{"plan", "code", "tests", "docs", "unknown", ""} {
		if got := DelegationContract(step); got != "" {
			t.Fatalf("DelegationContract(%q) = %q, want empty", step, got)
		}
	}
}
