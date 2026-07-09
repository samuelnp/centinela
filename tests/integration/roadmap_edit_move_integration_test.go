package integration_test

// Acceptance: specs/roadmap-edit-move.feature
// Scenario: move is allowed for a feature that another feature depends on (dependency is by name, not phase)
// Scenario: move preserves quality entries for an already-promoted feature
// Scenario: move preserves the feature's draft status and quality entries

import (
	"bytes"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const imBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},` +
	`{"name":"checkout-ui","dependsOn":["auth-service"]}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]}]}`

// TestEM_MoveKeepsDependencyValid moves a depended-on feature across phases and
// confirms the dependency (by name, not phase) still validates.
func TestEM_MoveKeepsDependencyValid(t *testing.T) {
	intoProject(t, imBody)
	if err := roadmap.Move(roadmap.RoadmapFile, roadmap.MoveRequest{
		Slug: "auth-service", ToPhase: "Phase 2: Growth"}); err != nil {
		t.Fatalf("Move: %v", err)
	}
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := roadmap.ValidateDependencies(r); err != nil {
		t.Fatalf("validate must PASS after move: %v", err)
	}
	for _, p := range r.Phases {
		for _, f := range p.Features {
			if f.Name == "checkout-ui" && (len(f.DependsOn) != 1 || f.DependsOn[0] != "auth-service") {
				t.Fatalf("dependency must survive the move: %+v", f)
			}
		}
	}
}

// TestEM_MovePreservesQuality moves a promoted feature and confirms the name-keyed
// quality entry file is left byte-identical.
func TestEM_MovePreservesQuality(t *testing.T) {
	intoProject(t, imBody)
	quality := `{"role":"roadmap-quality-evaluator","threshold":9,"features":[` +
		`{"name":"auth-service","scores":{"acceptanceCriteria":9,"userValue":9,` +
		`"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}]}`
	if err := os.WriteFile(roadmap.RoadmapQualityFile, []byte(quality), 0o644); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(roadmap.RoadmapQualityFile)
	if err := roadmap.Move(roadmap.RoadmapFile, roadmap.MoveRequest{
		Slug: "auth-service", ToPhase: "Phase 2: Growth"}); err != nil {
		t.Fatalf("Move: %v", err)
	}
	after, _ := os.ReadFile(roadmap.RoadmapQualityFile)
	if !bytes.Equal(before, after) {
		t.Fatal("move must not touch roadmap-quality.json")
	}
	r, _ := roadmap.Load()
	for _, p := range r.Phases {
		if p.Name == "Phase 2: Growth" {
			for _, f := range p.Features {
				if f.Name == "auth-service" {
					return
				}
			}
		}
	}
	t.Fatal("auth-service must now be in Phase 2")
}
