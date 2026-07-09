package acceptance_test

// Acceptance: specs/roadmap-edit-move.feature
// Scenario: move relocates a feature to the target phase, appending by default
// Scenario Outline: move --before/--after anchors the feature at the first, last, or middle position
// Scenario Outline: move refuses Backlog/Baseline as source or target, and unknown phase/anchor
// Scenario: reorder repositions a feature within its own phase
// Scenario: a no-op reorder leaves the file byte-identical

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const opsBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"checkout-ui"},{"name":"extra"}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"},{"name":"reporting"}]},` +
	`{"name":"Backlog","features":[{"name":"legacy-finding","summary":"s"}]}]}`

// emOrder returns the feature-name order of phase in project d.
func emOrder(t *testing.T, d, phase string) (out []string) {
	t.Helper()
	b, _ := os.ReadFile(emPath(d))
	var r roadmap.Roadmap
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, p := range r.Phases {
		if p.Name == phase {
			for _, f := range p.Features {
				out = append(out, f.Name)
			}
		}
	}
	return out
}

// TestAcc_MoveAppendAndAnchor drives an append then a --before anchor.
func TestAcc_MoveAppendAndAnchor(t *testing.T) {
	d := rmcProject(t, opsBody)
	if _, _, code := rmcRun(t, d, "roadmap", "move", "checkout-ui", "--to-phase", "Phase 2: Growth"); code != 0 {
		t.Fatalf("move append exit=%d", code)
	}
	if got := strings.Join(emOrder(t, d, "Phase 2: Growth"), ","); got != "billing-api,reporting,checkout-ui" {
		t.Fatalf("append order: %s", got)
	}
	d2 := rmcProject(t, opsBody)
	if _, _, code := rmcRun(t, d2, "roadmap", "move", "checkout-ui", "--to-phase", "Phase 2: Growth", "--before", "billing-api"); code != 0 {
		t.Fatalf("move anchor exit=%d", code)
	}
	if got := strings.Join(emOrder(t, d2, "Phase 2: Growth"), ","); got != "checkout-ui,billing-api,reporting" {
		t.Fatalf("anchor order: %s", got)
	}
}

// TestAcc_ReorderWithinAndNoOp repositions in-phase, then asserts a no-op is
// byte-identical.
func TestAcc_ReorderWithinAndNoOp(t *testing.T) {
	d := rmcProject(t, opsBody)
	if _, _, code := rmcRun(t, d, "roadmap", "reorder", "extra", "--before", "checkout-ui"); code != 0 {
		t.Fatalf("reorder exit=%d", code)
	}
	if got := strings.Join(emOrder(t, d, "Phase 1: Foundations"), ","); got != "extra,checkout-ui" {
		t.Fatalf("reorder order: %s", got)
	}
	d2 := rmcProject(t, opsBody)
	before, _ := os.ReadFile(emPath(d2))
	if _, _, code := rmcRun(t, d2, "roadmap", "reorder", "reporting", "--after", "billing-api"); code != 0 {
		t.Fatalf("no-op reorder exit=%d", code)
	}
	after, _ := os.ReadFile(emPath(d2))
	if string(before) != string(after) {
		t.Fatal("no-op reorder must be byte-identical")
	}
}

// TestAcc_MoveRefusalsByteIdentical drives guard rejections through the binary.
func TestAcc_MoveRefusalsByteIdentical(t *testing.T) {
	rows := [][]string{
		{"roadmap", "move", "checkout-ui", "--to-phase", "Backlog"},
		{"roadmap", "move", "checkout-ui", "--to-phase", "Phase 9: Nonexistent"},
		{"roadmap", "move", "checkout-ui", "--to-phase", "Phase 2: Growth", "--before", "ghost-anchor"},
		{"roadmap", "move", "legacy-finding", "--to-phase", "Phase 2: Growth"},
	}
	for _, args := range rows {
		d := rmcProject(t, opsBody)
		before, _ := os.ReadFile(emPath(d))
		if _, _, code := rmcRun(t, d, args...); code == 0 {
			t.Fatalf("row %v must be rejected", args)
		}
		after, _ := os.ReadFile(emPath(d))
		if string(before) != string(after) {
			t.Fatalf("row %v must be byte-identical", args)
		}
	}
}
