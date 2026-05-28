package evidence

import (
	"errors"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func swapOrchValidate(t *testing.T, fn func(path, feature, step string, role Role, uiPaths []string) error) {
	t.Helper()
	prev := orchValidate
	orchValidate = fn
	t.Cleanup(func() { orchValidate = prev })
}

func TestValidateFeatureEmitsFixHints(t *testing.T) {
	chdirToTemp(t)
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := WriteAtomic("alpha", orchestration.RoleBigThinker, s); err != nil {
		t.Fatal(err)
	}
	swapOrchValidate(t, func(_, _, _ string, _ Role, _ []string) error {
		return errors.New("edgeCases required in: x")
	})
	hints := ValidateFeature("alpha", nil)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint, got %d", len(hints))
	}
	if !strings.Contains(hints[0].Command, "append alpha big-thinker edgeCases") {
		t.Fatalf("hint missing append cmd: %s", hints[0].Command)
	}
}

func TestValidateNilOrchestrationError(t *testing.T) {
	swapOrchValidate(t, func(_, _, _ string, _ Role, _ []string) error {
		return nil
	})
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if errs := s.Validate("p", nil); errs != nil {
		t.Fatalf("expected nil, got %+v", errs)
	}
}

func TestValidateFeatureNoFilesNoHints(t *testing.T) {
	chdirToTemp(t)
	if hints := ValidateFeature("ghost", nil); len(hints) != 0 {
		t.Fatalf("unexpected hints: %+v", hints)
	}
}

func TestRoleFromPathIgnoresUnknown(t *testing.T) {
	if _, ok := roleFromPath(".workflow/alpha-mystery.json", "alpha"); ok {
		t.Fatal("should reject unknown role")
	}
	if r, ok := roleFromPath(".workflow/alpha-big-thinker.json", "alpha"); !ok || r != orchestration.RoleBigThinker {
		t.Fatalf("expected big-thinker, got %v ok=%v", r, ok)
	}
}

func TestGuessFieldKnownPatterns(t *testing.T) {
	cases := map[string]string{
		"missing feature-doc snapshot inputs: x": "inputs",
		"actionable outputs must be real files":  "outputs",
		"edgeCases required in: y":               "edgeCases",
		"invalid generatedAt in: z":              "generatedAt",
		"incomplete evidence fields: q":          "incomplete",
	}
	for msg, want := range cases {
		if got := guessField(msg); got != want {
			t.Errorf("guessField(%q) = %q, want %q", msg, got, want)
		}
	}
}
