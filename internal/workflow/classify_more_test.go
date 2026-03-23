package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestClassifyRoadmapArtifactsAndDefaults(t *testing.T) {
	cfg := &config.Config{}
	if ClassifyFile("/p/ROADMAP.md", cfg) != TypeRoadmap {
		t.Fatal("ROADMAP.md should be roadmap type")
	}
	if ClassifyFile("/p/.workflow/roadmap.json", cfg) != TypeRoadmap {
		t.Fatal("roadmap.json should be roadmap type")
	}
	if ClassifyFile("/p/internal/x.go", cfg) != TypeCode {
		t.Fatal("default code dirs should classify internal as code")
	}
}
