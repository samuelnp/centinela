package planadvisor

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestDirectiveAsksOnlyMissingQuestions(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                            //nolint:errcheck
	os.Chdir(d)                                                                                                  //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                           //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("surface: user-facing\n## Problem\ntext\n## Scope\ntext\n"), 0644) //nolint:errcheck
	out := Directive("f", &config.Config{})
	if strings.Contains(out, "What exact user or operator pain") {
		t.Fatalf("did not expect repeated problem question: %s", out)
	}
	if !strings.Contains(out, "observable behaviors") || !strings.Contains(out, "loading, empty, and error") {
		t.Fatalf("expected missing feature-specialist questions, got: %s", out)
	}
	if strings.Count(out, "- [") > config.DefaultPlanQuestionCap {
		t.Fatalf("expected at most %d questions, got: %s", config.DefaultPlanQuestionCap, out)
	}
	if !strings.Contains(out, "mobile-first") {
		t.Fatalf("expected user-facing mobile-first guidance, got: %s", out)
	}
}
