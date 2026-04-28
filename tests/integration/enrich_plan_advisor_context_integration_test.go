package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

func TestPlanAdvisorUsesDependencyAndQualityContext(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     //nolint:errcheck
	os.Chdir(d)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("## Problem\ntext\n"), 0644)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"dep"},{"name":"sib"},{"name":"f"}]}]}`), 0644)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[{"name":"dep","dependsOn":[]},{"name":"sib","dependsOn":[]},{"name":"f","dependsOn":["dep"]}]}`), 0644)                                                                                                                                                                                                                                                                                                                                                                                                                                          //nolint:errcheck
	os.WriteFile(".workflow/roadmap-quality.json", []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"dep","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"integration assumptions need clarity"},{"name":"sib","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"ok"},{"name":"f","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"acceptance clarity low"}]}`), 0644) //nolint:errcheck
	out := planadvisor.Directive("f", &config.Config{})
	if !strings.Contains(out, "dependencies first: dep") || !strings.Contains(out, "roadmap quality notes: dep: integration assumptions need clarity") {
		t.Fatalf("expected dependency and quality context, got: %s", out)
	}
	if !strings.Contains(out, "shared-contract constraints") || !strings.Contains(out, "acceptance criteria should close that clarity gap") {
		t.Fatalf("expected context-aware questions, got: %s", out)
	}
}
