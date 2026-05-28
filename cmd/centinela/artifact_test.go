package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
)

func TestArtifactNewHappyPath(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	artifactForce = false
	if err := runArtifactNew(nil, []string{"alpha", "edge-cases"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(".workflow/alpha-edge-cases.md"); err != nil {
		t.Fatalf("artifact not written: %v", err)
	}
}

func TestArtifactNewUnknownKind(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	err := runArtifactNew(nil, []string{"alpha", "bogus"})
	if err == nil {
		t.Fatal("expected unknown-kind error")
	}
	if !strings.Contains(err.Error(), "edge-cases") {
		t.Fatalf("error missing allowed-kinds list: %v", err)
	}
}

func TestArtifactNewRefusesWithoutForce(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	artifactForce = false
	if err := runArtifactNew(nil, []string{"alpha", "gatekeeper"}); err != nil {
		t.Fatal(err)
	}
	err := runArtifactNew(nil, []string{"alpha", "gatekeeper"})
	if !errors.Is(err, evidence.ErrArtifactExists) {
		t.Fatalf("expected ErrArtifactExists, got %v", err)
	}
	artifactForce = true
	t.Cleanup(func() { artifactForce = false })
	if err := runArtifactNew(nil, []string{"alpha", "gatekeeper"}); err != nil {
		t.Fatalf("force should succeed: %v", err)
	}
}

func TestArtifactNewUnknownFeature(t *testing.T) {
	chdirEvidenceTemp(t)
	err := runArtifactNew(nil, []string{"ghost", "edge-cases"})
	if err == nil || !strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("expected unknown-feature error, got %v", err)
	}
}
