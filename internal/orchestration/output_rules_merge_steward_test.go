package orchestration

import (
	"os"
	"strings"
	"testing"
)

func TestValidateActionableOutputs_MergeSteward(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	_ = os.MkdirAll(".workflow", 0755)
	_ = os.MkdirAll("internal", 0755)
	_ = os.WriteFile("internal/foo.go", []byte("package internal"), 0644)
	_ = os.WriteFile(".workflow/alpha-merge-steward.md", []byte("ok"), 0644)

	// Missing the required report path -> error citing merge-steward outputs.
	err := validateActionableOutputs("x", "alpha", RoleMergeSteward, []string{"internal/foo.go"}, nil)
	if err == nil || !strings.Contains(err.Error(), "merge-steward outputs must include") {
		t.Fatalf("expected merge-steward error, got %v", err)
	}

	// Report present -> success.
	if err := validateActionableOutputs("x", "alpha", RoleMergeSteward,
		[]string{".workflow/alpha-merge-steward.md"}, nil); err != nil {
		t.Fatalf("expected merge-steward success, got %v", err)
	}
}
