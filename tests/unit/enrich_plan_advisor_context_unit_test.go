package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

func TestPlanAdvisorSummarizesRelatedContextWithoutRawDump(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                                                                                            //nolint:errcheck
	os.Chdir(d)                                                                                                                                                                                                  //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                                                                                                           //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                                                                                               //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("## Problem\ntext\n"), 0644)                                                                                                                                       //nolint:errcheck
	os.WriteFile("docs/features/dep.md", []byte("FULL RAW PARAGRAPH SHOULD NOT APPEAR"), 0644)                                                                                                                   //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"dep"},{"name":"sib"},{"name":"f"}]}]}`), 0644)                                                                   //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[{"name":"dep","dependsOn":[]},{"name":"sib","dependsOn":[]},{"name":"f","dependsOn":["dep"]}]}`), 0644) //nolint:errcheck
	os.WriteFile(".workflow/dep-edge-cases.md", []byte("- duplicate webhook retries"), 0644)                                                                                                                     //nolint:errcheck
	out := planadvisor.Directive("f", &config.Config{})
	if !strings.Contains(out, "dependencies first: dep") || !strings.Contains(out, "same-phase siblings: sib") {
		t.Fatalf("expected summarized related context, got: %s", out)
	}
	if strings.Contains(out, "FULL RAW PARAGRAPH SHOULD NOT APPEAR") {
		t.Fatalf("expected no raw file dump, got: %s", out)
	}
}
