package evidence

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestFixHintStringHasContext(t *testing.T) {
	h := FixHint{
		Feature: "alpha", Role: orchestration.RoleBigThinker,
		Field: "outputs", Message: "missing",
		Command: "centinela evidence append alpha big-thinker outputs <value>",
	}
	got := h.String()
	if !strings.Contains(got, "[alpha/big-thinker]") || !strings.Contains(got, "fix:") {
		t.Fatalf("hint render: %q", got)
	}
}

func TestSuggestCommandPerField(t *testing.T) {
	cases := map[string]string{
		"inputs":      "append",
		"outputs":     "append",
		"edgeCases":   "append",
		"mobileFirst": "set",
		"status":      "set",
		"generatedAt": "set",
		"feature":     "init",
	}
	for field, verb := range cases {
		got := suggestCommand("alpha", orchestration.RoleBigThinker, FieldError{Field: field, Message: "x"})
		if !strings.Contains(got, verb) {
			t.Errorf("field=%s wanted verb %q in %q", field, verb, got)
		}
	}
}

func TestSuggestCommandMissingEvidence(t *testing.T) {
	got := suggestCommand("alpha", orchestration.RoleBigThinker, FieldError{Message: "missing evidence json"})
	if !strings.Contains(got, "init") {
		t.Fatalf("missing-evidence hint should suggest init, got %q", got)
	}
}

func TestSuggestCommandUnknownReturnsEmpty(t *testing.T) {
	if got := suggestCommand("alpha", orchestration.RoleBigThinker, FieldError{Message: "no clue"}); got != "" {
		t.Fatalf("expected empty for unknown, got %q", got)
	}
}

func TestFixHintNoCommandStillRenders(t *testing.T) {
	h := FixHint{Feature: "alpha", Role: orchestration.RoleBigThinker, Message: "boom"}
	got := h.String()
	if strings.Contains(got, "fix:") {
		t.Fatalf("no command but rendered fix: %q", got)
	}
}
