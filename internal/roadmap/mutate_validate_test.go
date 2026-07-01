package roadmap

import (
	"strings"
	"testing"
)

// TestRequirePlannedStatus passes for planned and rejects in-progress/done.
func TestRequirePlannedStatus(t *testing.T) {
	crudChdir(t, crudBody) // no workflow state → FeatureStatus == planned
	if err := requirePlannedStatus("lonely-feature"); err != nil {
		t.Fatalf("planned feature must pass: %v", err)
	}
	seedStatus(t, "lonely-feature", "code")
	err := requirePlannedStatus("lonely-feature")
	if err == nil || !strings.Contains(err.Error(), "in-progress") {
		t.Fatalf("in-progress must be refused, got %v", err)
	}
}

// TestRequireNoDependents rejects a depended-on feature and passes otherwise.
func TestRequireNoDependents(t *testing.T) {
	doc := docFrom(t, crudBody)
	err := doc.requireNoDependents("auth-service")
	if err == nil || !strings.Contains(err.Error(), "checkout-ui") {
		t.Fatalf("depended feature must be refused naming dependent, got %v", err)
	}
	if err := doc.requireNoDependents("checkout-ui"); err != nil {
		t.Fatalf("undepended feature must pass: %v", err)
	}
}

// TestJoinNames renders a comma-separated quoted list.
func TestJoinNames(t *testing.T) {
	if got := joinNames([]string{"a", "b"}); got != `"a", "b"` {
		t.Fatalf("joinNames = %q", got)
	}
	if got := joinNames(nil); got != "" {
		t.Fatalf("empty joinNames = %q", got)
	}
}
