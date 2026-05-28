package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// checkDependencyGuard returns an error naming the feature and every unmet dep
// when a dependency is not done, and nil when all deps are done or there are none.
func TestCheckDependencyGuard_BlockedAndAllowed(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck

	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P", Features: []roadmap.Feature{
		{Name: "dep-a"},
		{Name: "dep-b"},
		{Name: "feature-x", DependsOn: []string{"dep-a", "dep-b"}},
		{Name: "free"},
	}}}}

	// Both deps planned → blocked, error names feature + both deps.
	err := checkDependencyGuard(r, "feature-x")
	if err == nil {
		t.Fatal("expected blocked error for feature-x")
	}
	msg := err.Error()
	for _, want := range []string{"feature-x", "dep-a", "dep-b"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("guard error should name %q, got: %s", want, msg)
		}
	}

	// Feature with no deps → allowed.
	if err := checkDependencyGuard(r, "free"); err != nil {
		t.Fatalf("no-dep feature should be allowed, got: %v", err)
	}

	// Both deps done → allowed.
	seedWF(t, "dep-a", "done")
	seedWF(t, "dep-b", "done")
	if err := checkDependencyGuard(r, "feature-x"); err != nil {
		t.Fatalf("all-deps-done should be allowed, got: %v", err)
	}
}
