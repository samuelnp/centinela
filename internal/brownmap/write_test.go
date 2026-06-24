package brownmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func samplePlan() Plan {
	return Plan{Roadmap: roadmap.Roadmap{Intro: "i", Phases: []roadmap.Phase{
		{Name: roadmap.BaselinePhaseName, Features: []roadmap.Feature{{Name: "a"}}},
	}}, BaselineCount: 1}
}

func TestWriteDraft_WritesDraftPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "roadmap.brownfield.json")
	wrote, err := WriteDraft(path, samplePlan())
	if err != nil || wrote != path {
		t.Fatalf("WriteDraft err=%v wrote=%q", err, wrote)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("draft file not written: %v", err)
	}
}

func TestWriteDraft_RefusesCanonicalRoadmap(t *testing.T) {
	_, err := WriteDraft(roadmap.RoadmapFile, samplePlan())
	if err == nil {
		t.Fatal("WriteDraft must refuse the canonical roadmap path")
	}
	if _, statErr := os.Stat(roadmap.RoadmapFile); statErr == nil {
		t.Fatal("refused write must not create the canonical file")
	}
}

func TestWriteDraft_Deterministic(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json")
	b := filepath.Join(dir, "b.json")
	if _, err := WriteDraft(a, samplePlan()); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteDraft(b, samplePlan()); err != nil {
		t.Fatal(err)
	}
	ba, _ := os.ReadFile(a)
	bb, _ := os.ReadFile(b)
	if string(ba) != string(bb) || len(ba) == 0 {
		t.Fatal("draft output must be byte-identical for the same plan")
	}
}

func TestWriteDraft_MkdirFailureWrapped(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// path under a regular file forces MkdirAll to fail.
	if _, err := WriteDraft(filepath.Join(blocker, "nested.json"), samplePlan()); err == nil {
		t.Fatal("expected an error when the parent dir cannot be created")
	}
}
