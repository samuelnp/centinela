package unit_test

// Acceptance: specs/roadmap-crud-add-remove.feature
// Scenario: add creates a draft in a chosen schedulable phase and validate stays PASS
// Scenario Outline: add rejects invalid input and leaves roadmap.json byte-identical
// Scenario: remove deletes a planned feature and leaves the file valid
// Scenario: a freshly-added draft simultaneously satisfies all four draft readers

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const unitBody = `{"phases":[{"name":"Phase 1","features":[{"name":"auth-service"}]},` +
	`{"name":"Phase 2","features":[{"name":"billing-api"}]}]}`

func unitPath(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "roadmap.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func loadPath(t *testing.T, p string) *roadmap.Roadmap {
	t.Helper()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var r roadmap.Roadmap
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return &r
}

// TestAddSetsDraftAndValidates confirms an added feature is a draft that keeps
// dependency validation PASS.
func TestAddSetsDraftAndValidates(t *testing.T) {
	p := unitPath(t, unitBody)
	if err := roadmap.Add(p, roadmap.AddRequest{Slug: "new-widget", Phase: "Phase 1"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	r := loadPath(t, p)
	if !roadmap.IsDraftFeature(r, "new-widget") {
		t.Fatal("added feature must be a draft")
	}
	if err := roadmap.ValidateDependencies(r); err != nil {
		t.Fatalf("validate must stay PASS: %v", err)
	}
}

// TestAddRejectionByteIdentical confirms a rejected add writes nothing.
func TestAddRejectionByteIdentical(t *testing.T) {
	p := unitPath(t, unitBody)
	before, _ := os.ReadFile(p)
	if err := roadmap.Add(p, roadmap.AddRequest{Slug: "auth-service", Phase: "Phase 1"}); err == nil {
		t.Fatal("duplicate slug must be rejected")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected add must leave roadmap.json byte-identical")
	}
}

// TestFourReaderInvariant confirms all four readers agree a draft is exempt.
func TestFourReaderInvariant(t *testing.T) {
	p := unitPath(t, unitBody)
	if err := roadmap.Add(p, roadmap.AddRequest{Slug: "widget", Phase: "Phase 1"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	r := loadPath(t, p)
	if roadmap.NonBacklogFeatureSet(r)["widget"] { // reader 1
		t.Fatal("coverage set must exempt the draft")
	}
	for _, name := range roadmap.ReadySet(r) { // reader 2
		if name == "widget" {
			t.Fatal("draft must not be ready")
		}
	}
	planned, _, _ := r.Summary() // reader 3
	if planned != 2 {            // auth-service + billing-api only
		t.Fatalf("draft must not be counted: planned=%d", planned)
	}
	for _, ph := range roadmap.BuildView(r).Phases { // reader 4
		for _, f := range ph.Features {
			if f.Name == "widget" && (!f.Draft || f.Readiness != "draft") {
				t.Fatalf("view reader must flag the draft: %+v", f)
			}
		}
	}
}
