package reconstruct

import (
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
)

func TestPromote_RoleHints(t *testing.T) {
	s := newSignals(analyze.Inventory{})
	cases := []struct {
		pkg  string
		role Role
		ok   bool
	}{
		{"cmd/app", RoleCommand, true},
		{"internal/command", RoleCommand, true},
		{"internal/handler", RoleEndpoint, true},
		{"app/controllers", RoleEndpoint, true},
		{"src/routes", RoleEndpoint, true},
		{"pkg/api/v1", RoleEndpoint, true},
		{"internal/service", RoleModule, true},
		{"domain/order", RoleModule, true},
		{"app/usecase", RoleModule, true},
		{"pkg/use_case", RoleModule, true},
		{"lib/core", RoleModule, true},
		{"leaf", "", false},
	}
	for _, c := range cases {
		role, _, ok := promote(c.pkg, s)
		if ok != c.ok || role != c.role {
			t.Errorf("promote(%q) = (%q,%v), want (%q,%v)", c.pkg, role, ok, c.role, c.ok)
		}
	}
}

func TestPromote_FrameworkSignals(t *testing.T) {
	cobra := newSignals(analyze.Inventory{Manifests: []analyze.Manifest{{Framework: "cobra"}}})
	if role, _, ok := promote("leaf", cobra); !ok || role != RoleCommand {
		t.Fatalf("cobra framework must promote leaf as command, got (%q,%v)", role, ok)
	}
	gin := newSignals(analyze.Inventory{Manifests: []analyze.Manifest{{Framework: "gin"}}})
	if role, _, ok := promote("leaf", gin); !ok || role != RoleEndpoint {
		t.Fatalf("gin framework must promote leaf as endpoint, got (%q,%v)", role, ok)
	}
}

func TestExcluded_AllRules(t *testing.T) {
	for _, p := range []string{"x_test", "tests/x", "spec/y", "a.pb.go", "node_modules/q",
		"vendor/q", "dist/q", "x/gen/y", "generated/z", "internal/config", "pkg/mocks", "test/fixtures"} {
		if !excluded(p) {
			t.Errorf("expected %q excluded", p)
		}
	}
	if excluded("internal/service") {
		t.Error("internal/service must not be excluded")
	}
}
