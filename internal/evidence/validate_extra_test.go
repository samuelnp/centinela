package evidence

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestValidateFeatureIgnoresUnknownRoleFiles(t *testing.T) {
	chdirToTemp(t)
	if err := os.WriteFile(".workflow/alpha-mystery.json", []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if hints := ValidateFeature("alpha", nil); len(hints) != 0 {
		t.Fatalf("expected unknown-role files ignored, got %+v", hints)
	}
}

func TestValidateFeatureValidatorReturnsNoErrors(t *testing.T) {
	chdirToTemp(t)
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := WriteAtomic("alpha", orchestration.RoleBigThinker, s); err != nil {
		t.Fatal(err)
	}
	swapOrchValidate(t, func(_, _, _ string, _ Role, _ []string) error { return nil })
	if hints := ValidateFeature("alpha", nil); len(hints) != 0 {
		t.Fatalf("expected no hints, got %+v", hints)
	}
}

func TestGuessFieldUnknownReturnsEmpty(t *testing.T) {
	if got := guessField("totally unrelated"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestValidateFeatureWithReadErrorEmitsInitHint(t *testing.T) {
	chdirToTemp(t)
	// Create a file that looks like a role file but with garbage content,
	// so Read fails inside hintsForFile.
	if err := os.WriteFile(".workflow/alpha-big-thinker.json", []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	hints := ValidateFeature("alpha", nil)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint, got %d", len(hints))
	}
	if hints[0].Command == "" {
		t.Fatal("expected non-empty command")
	}
}
