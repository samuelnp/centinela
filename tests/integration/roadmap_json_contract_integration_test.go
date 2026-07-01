package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// Load-from-disk → BuildView excludes non-schedulable phases and marshals
// byte-identically across repeated builds (the determinism contract).
func TestRoadmapViewFromDiskExcludesNonSchedulableAndIsStable(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phases":[{"name":"Baseline","features":[{"name":"legacy"}]},` +
		`{"name":"Q1","features":[{"name":"a"},{"name":"b","dependsOn":["a"]}]}]}`
	if err := os.WriteFile(filepath.Join(".workflow", "roadmap.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	first, err := json.MarshalIndent(roadmap.BuildView(r), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.MarshalIndent(roadmap.BuildView(r), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("BuildView must be byte-stable:\n%s\n---\n%s", first, second)
	}
	var v roadmap.RoadmapView
	if err := json.Unmarshal(first, &v); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(v.Phases) != 1 || v.Phases[0].Name != "Q1" {
		t.Fatalf("Baseline must be excluded from the view: %+v", v.Phases)
	}
}
