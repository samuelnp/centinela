package acceptance_test

// Acceptance: specs/roadmap-crud-add-remove.feature
// Scenario: remove deletes a planned feature and leaves the file valid
// Scenario: rm is an alias for remove
// Scenario: remove a feature that does not exist errors "not found"
// Scenario: promote finalizes a draft in place — no phase move, draft cleared, artifacts written

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestAcc_RemoveAndAlias removes an added draft via both remove and its rm alias.
func TestAcc_RemoveAndAlias(t *testing.T) {
	for _, verb := range []string{"remove", "rm"} {
		d := crudSeeded(t)
		if _, _, code := rmcRun(t, d, "roadmap", "add", "old-widget", "--phase", "Phase 1: Foundations"); code != 0 {
			t.Fatalf("add exit=%d", code)
		}
		if _, _, code := rmcRun(t, d, "roadmap", verb, "old-widget"); code != 0 {
			t.Fatalf("%s exit=%d", verb, code)
		}
		data, _ := os.ReadFile(filepath.Join(d, ".workflow", "roadmap.json"))
		if strings.Contains(string(data), "old-widget") {
			t.Fatalf("%s must delete old-widget", verb)
		}
		var r roadmap.Roadmap
		if err := json.Unmarshal(data, &r); err != nil {
			t.Fatalf("result must be valid JSON: %v", err)
		}
	}
}

// TestAcc_RemoveNotFound errors and leaves the file byte-identical.
func TestAcc_RemoveNotFound(t *testing.T) {
	d := crudSeeded(t)
	p := filepath.Join(d, ".workflow", "roadmap.json")
	before, _ := os.ReadFile(p)
	_, stderr, code := rmcRun(t, d, "roadmap", "remove", "ghost-feature")
	if code == 0 || !strings.Contains(stderr, "not found") {
		t.Fatalf("remove ghost must error not-found: code=%d err=%s", code, stderr)
	}
	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Fatal("failed remove must be byte-identical")
	}
}

// TestAcc_PromoteDraftInPlace finalizes a draft in place and keeps validate PASS.
func TestAcc_PromoteDraftInPlace(t *testing.T) {
	d := crudSeeded(t)
	if _, _, code := rmcRun(t, d, "roadmap", "add", "new-widget", "--phase", "Phase 1: Foundations"); code != 0 {
		t.Fatalf("add exit=%d", code)
	}
	if _, _, code := rmcRun(t, d, "roadmap", "promote", "new-widget", "--scores", "9,9,9,9,9,9"); code != 0 {
		t.Fatalf("promote in place exit=%d", code)
	}
	data, _ := os.ReadFile(filepath.Join(d, ".workflow", "roadmap.json"))
	if strings.Contains(string(data), `"draft":true`) {
		t.Fatal("draft flag must be cleared by in-place finalize")
	}
	if !strings.Contains(string(data), "new-widget") {
		t.Fatal("feature must remain in place (no move)")
	}
	if _, _, code := rmcRun(t, d, "roadmap", "validate"); code != 0 {
		t.Fatalf("validate must PASS after finalize, exit=%d", code)
	}
	// analysis gained the finalized feature.
	a, _ := os.ReadFile(filepath.Join(d, ".workflow", "roadmap-analysis.json"))
	if !strings.Contains(string(a), "new-widget") {
		t.Fatal("analysis must gain the finalized feature")
	}
}
