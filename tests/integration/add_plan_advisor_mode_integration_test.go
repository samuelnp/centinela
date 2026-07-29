package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

func TestPlanAdvisorAsksUserFacingMobileQuestionsWhenMissing(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                            //nolint:errcheck
	os.Chdir(d)                                                                                                  //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                           //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n## Problem\ntext\n## Scope\ntext\n"), 0644) //nolint:errcheck
	out := planadvisor.Directive("f", &config.Config{})
	if !strings.Contains(out, "mobile-first") || !strings.Contains(out, "loading, empty, and error") {
		t.Fatalf("expected user-facing UX questions, got: %s", out)
	}
	if !strings.Contains(out, "One planner agent, two lenses: strategy first, then spec.") {
		t.Fatalf("expected the one-agent two-lens header, got: %s", out)
	}
	if !strings.Contains(out, "[spec]") {
		t.Fatalf("expected spec lens tags on the UX questions, got: %s", out)
	}
	for _, retired := range []string{"[big-thinker]", "[feature-specialist]"} {
		if strings.Contains(out, retired) {
			t.Fatalf("advisor must not tag questions %s, got: %s", retired, out)
		}
	}
}
