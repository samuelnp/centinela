package reconstruct

import (
	"strings"
	"testing"
)

func TestBriefStub_ShapeAndTodos(t *testing.T) {
	body := briefStub(Target{Pkg: "internal/order", Slug: "internal-order", Role: RoleModule, Reason: "behavioral package"})
	for _, want := range []string{
		"# Feature: internal-order",
		"**Reconstructed from:** `internal/order` (module)",
		"**Selected because:** behavioral package",
		"## Problem", "## User value", "## What ships", "## Acceptance criteria",
		"specs/internal-order.feature",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("brief missing %q:\n%s", want, body)
		}
	}
	if got := strings.Count(body, todoMarker); got != 4 {
		t.Fatalf("expected 4 TODO markers in brief, got %d", got)
	}
}

func TestBriefStub_EmptyRoleNormalizes(t *testing.T) {
	body := briefStub(Target{Pkg: "p", Slug: "p", Role: "", Reason: "r"})
	if !strings.Contains(body, "(module)") {
		t.Fatalf("empty role must render as module:\n%s", body)
	}
}
