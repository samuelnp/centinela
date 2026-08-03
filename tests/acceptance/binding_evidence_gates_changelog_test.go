// Acceptance: specs/binding-evidence-gates.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/workflow"
)

// begDocsGate seeds an internal feature at the docs step with the given
// changelog body and returns the artifact gate's verdict.
func begDocsGate(t *testing.T, feature, changelog string) error {
	t.Helper()
	wf := begRepo(t, feature, false)
	wf.CurrentStep = "docs"
	begSave(t, wf)
	mustWrite(t, filepath.Join(workflow.WorkflowDir, feature+"-changelog.md"), changelog)
	return workflow.ValidateArtifacts(feature, "docs", begCfg())
}

// Scenario: An unfilled changelog template fails the docs gate
func TestBEG_UnfilledChangelogFailsDocsGate(t *testing.T) {
	err := begDocsGate(t, "demo", "- <FILL: type>: <FILL: one-line summary of the change>\n")
	if err == nil {
		t.Fatal("an unfilled changelog template must fail the docs gate")
	}
	if !strings.Contains(err.Error(), "<FILL: ...>") || !strings.Contains(err.Error(), "one-line summary") {
		t.Fatalf("error must say to replace the placeholder: %v", err)
	}
}

// TestBEG_ScaffoldedChangelogFailsItsOwnGate feeds the REAL scaffold the CLI
// writes straight into the REAL gate, so the renderer and the detector can
// never drift apart: if `artifact new <f> changelog` ever stops emitting a
// slot, or the gate stops seeing one, this fails.
func TestBEG_ScaffoldedChangelogFailsItsOwnGate(t *testing.T) {
	_, bodies, err := evidence.RenderTemplate(evidence.KindChangelog, "demo")
	if err != nil {
		t.Fatalf("render changelog scaffold: %v", err)
	}
	if err := begDocsGate(t, "demo", string(bodies[0])); err == nil {
		t.Fatal("the CLI's own changelog scaffold must fail the CLI's own docs gate")
	}
	filled := strings.NewReplacer(
		"<FILL: type>", "fix",
		"<FILL: one-line summary of the change>", "bind the evidence gates",
	).Replace(string(bodies[0]))
	if err := begDocsGate(t, "demo", filled); err != nil {
		t.Fatalf("the same scaffold with its slots replaced must pass: %v", err)
	}
}

// A stub behind a heading used to pass, because only the first non-blank line
// was inspected. It is refused now, and prose that quotes the marker in its
// generic citation form still passes — including the line this very feature
// ships, which the positional rule would have rejected as a false positive.
//
// Scenario: A changelog stub behind a heading fails the docs gate
func TestBEG_ChangelogStubBehindHeadingFailsDocsGate(t *testing.T) {
	stub := "- <FILL: type>: <FILL: one-line summary of the change>\n"
	if err := begDocsGate(t, "demo", "## Changelog\n\n"+stub); err == nil {
		t.Fatal("a heading must not re-open the gate on an untouched scaffold")
	}
	if err := begDocsGate(t, "demo", "- fix: <FILL: one-line summary of the change>\n"); err == nil {
		t.Fatal("a half-filled entry is still a template")
	}
	own := "- fix: reject an unreplaced <FILL: ...> marker behind a preamble\n"
	if err := begDocsGate(t, "demo", own); err != nil {
		t.Fatalf("this feature's own changelog line must pass its own gate: %v", err)
	}
}

// Scenario: A filled-in changelog passes the docs gate
func TestBEG_FilledChangelogPassesDocsGate(t *testing.T) {
	body := "- feat: bind the evidence gates so a stub can no longer pass\n"
	if err := begDocsGate(t, "demo", body); err != nil {
		t.Fatalf("a real one-line entry must pass the docs gate: %v", err)
	}
}
