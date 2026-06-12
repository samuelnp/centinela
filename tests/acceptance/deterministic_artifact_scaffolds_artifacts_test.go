package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
)

// writeGatekeeperArtifact replicates `centinela artifact new <feature> gatekeeper`
// and returns the rendered body.
func writeGatekeeperArtifact(t *testing.T) string {
	t.Helper()
	paths, err := evidence.WriteArtifact("demo", evidence.KindGatekeeper, true)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Gatekeeper artifact pre-fills Analyzed Specs from existing specs
func TestDAS_GatekeeperPreFillsAnalyzedSpecs(t *testing.T) {
	dasChdir(t)
	if err := os.MkdirAll("specs", 0o755); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"b.feature", "a.feature"} {
		if err := os.WriteFile("specs/"+s, []byte("Feature: x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	body := writeGatekeeperArtifact(t)
	ia, ib := strings.Index(body, "specs/a.feature"), strings.Index(body, "specs/b.feature")
	if ia < 0 || ib < 0 {
		t.Fatalf("Analyzed Specs missing entries: %s", body)
	}
	if ia > ib {
		t.Fatalf("Analyzed Specs not sorted: %s", body)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Gatekeeper artifact Analyzed Specs is an empty list when no specs exist
func TestDAS_GatekeeperEmptyAnalyzedSpecs(t *testing.T) {
	dasChdir(t)
	body := writeGatekeeperArtifact(t)
	section := body[strings.Index(body, "Analyzed Specs"):]
	section = section[:strings.Index(section, "Findings")]
	if strings.Contains(section, "specs/") {
		t.Fatalf("Analyzed Specs should list no spec files: %s", body)
	}
	if !strings.Contains(section, "<FILL:") {
		t.Fatalf("empty Analyzed Specs should carry a single fill row: %s", body)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Artifact bodies use FILL slots for substance sections
func TestDAS_ArtifactBodiesUseFillSlots(t *testing.T) {
	dasChdir(t)
	body := writeGatekeeperArtifact(t)
	if !hasFill(body) {
		t.Fatalf("gatekeeper substance sections missing fill slots: %s", body)
	}
}

// Acceptance: specs/deterministic-artifact-scaffolds.feature
// Scenario: Artifact Status and Date lines stay parseable by validate
func TestDAS_ArtifactStatusDateLinesIntact(t *testing.T) {
	dasChdir(t)
	body := writeGatekeeperArtifact(t)
	for _, want := range []string{"**Status:**", "**Date:**"} {
		if !strings.Contains(body, want) {
			t.Fatalf("gatekeeper missing %q line: %s", want, body)
		}
	}
}
