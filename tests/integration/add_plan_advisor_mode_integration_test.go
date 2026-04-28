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
	if !strings.Contains(out, "big-thinker") || !strings.Contains(out, "feature-specialist") {
		t.Fatalf("expected both advisor lenses, got: %s", out)
	}
}
