package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

func TestPlanAdvisorAvoidsRepeatingCoveredTopics(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                            //nolint:errcheck
	os.Chdir(d)                                                                                                                  //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                           //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                                                              //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                                                                   //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("## Problem\ntext\n## Scope\ntext\n## Constraints\ntext\n## Risks\ntext\n"), 0644) //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: x\nScenario: y\nGiven a\nWhen b\nThen c\n"), 0644)                          //nolint:errcheck
	out := planadvisor.Directive("f", &config.Config{})
	if strings.Contains(out, "What exact user or operator pain") || strings.Contains(out, "What constraints or non-negotiables") {
		t.Fatalf("expected covered strategic topics to be suppressed, got: %s", out)
	}
	if strings.Count(out, "- [") > 4 {
		t.Fatalf("expected at most 4 questions, got: %s", out)
	}
}
