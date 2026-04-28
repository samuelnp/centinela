package planadvisor

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestRoadmapContextFallbackBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if got := dependencyNames("f"); got != nil {
		t.Fatalf("expected nil dependencies without analysis, got %v", got)
	}
	os.MkdirAll(".workflow", 0755)                                                                                                                         //nolint:errcheck
	os.WriteFile(roadmap.RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[{"name":"f","dependsOn":["dep"]}]}`), 0644)             //nolint:errcheck
	os.WriteFile(roadmap.RoadmapFile, []byte(`{"phases":[{"name":"P1","features":[{"name":"dep"},{"name":"sib1"},{"name":"sib2"},{"name":"f"}]}]}`), 0644) //nolint:errcheck
	deps := dependencyNames("f")
	if len(deps) != 1 || deps[0] != "dep" {
		t.Fatalf("unexpected dependencies: %v", deps)
	}
	sibs := siblingNames("f", deps)
	if len(sibs) != 2 || sibs[0] != "sib1" || sibs[1] != "sib2" {
		t.Fatalf("unexpected siblings: %v", sibs)
	}
	if got := siblingNames("missing", nil); got != nil {
		t.Fatalf("expected nil siblings for missing feature, got %v", got)
	}
}
