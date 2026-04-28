package planadvisor

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestRelatedQualityNotesAndHelpers(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint:errcheck
	os.Chdir(d)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  //nolint:errcheck
	os.WriteFile(roadmap.RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[{"name":"dep","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"integration assumptions need clarity"},{"name":"sib","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"ok"},{"name":"f","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"acceptance clarity low"}]}`), 0644) //nolint:errcheck
	notes := relatedQualityNotes("f", []string{"dep", "sib"})
	if len(notes) != 2 || notes[0] != "dep: integration assumptions need clarity" || notes[1] != "f: acceptance clarity low" {
		t.Fatalf("unexpected quality notes: %v", notes)
	}
	if got := relatedQualityNotes("f", []string{"missing"}); len(got) != 1 || got[0] != "f: acceptance clarity low" {
		t.Fatalf("expected current feature quality fallback, got %v", got)
	}
	os.Remove(roadmap.RoadmapQualityFile) //nolint:errcheck
	if got := relatedQualityNotes("f", []string{"dep"}); got != nil {
		t.Fatalf("expected nil without quality file, got %v", got)
	}
	if !phaseHasFeature(roadmap.Phase{Features: []roadmap.Feature{{Name: "f"}}}, "f") || phaseHasFeature(roadmap.Phase{}, "f") {
		t.Fatal("expected phaseHasFeature branch coverage")
	}
	if got := take([]string{"a", "b", "c"}, 2); len(got) != 2 || got[1] != "b" {
		t.Fatalf("unexpected take result: %v", got)
	}
}
