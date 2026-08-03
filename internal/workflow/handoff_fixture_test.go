package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// hoCase is one derivation row: a workflow shape plus the role whose successor
// the workflow's OWN contract must produce.
type hoCase struct {
	name           string
	userFacing     bool
	order          []string
	legacyValidate bool
	legacyPlan     bool
	step           string
	role           orchestration.Role
	want           string
}

// hoWrite creates dir and writes name inside it.
func hoWrite(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// hoRepo chdirs into a temp repo carrying tc's feature brief and saved
// workflow, so every derivation is read from real state rather than a stub.
func hoRepo(t *testing.T, feature string, tc hoCase) {
	t.Helper()
	t.Chdir(t.TempDir())
	brief := "# " + feature + "\n"
	if tc.userFacing {
		brief += "surface: user-facing\n"
	}
	hoWrite(t, "docs/features", feature+".md", brief)
	hoWrite(t, WorkflowDir, ".gitkeep", "")
	wf := New(feature)
	if tc.order != nil {
		wf.StepOrder = tc.order
	}
	if tc.legacyValidate {
		wf.ValidateContract = ""
	}
	if tc.legacyPlan {
		wf.PlanContract = ""
	}
	if err := Save(wf); err != nil {
		t.Fatal(err)
	}
}

// hoEvidence writes the one field the chain check reads.
func hoEvidence(t *testing.T, feature string, role orchestration.Role, handoffTo string) {
	t.Helper()
	body := fmt.Sprintf(`{"feature":%q,"role":%q,"handoffTo":%q}`, feature, role, handoffTo)
	if err := os.WriteFile(orchestration.JSONPath(feature, role), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// runHandoffMatrix asserts the derived successor for every row.
func runHandoffMatrix(t *testing.T, cases []hoCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hoRepo(t, "f", tc)
			got, ok := ExpectedHandoff("f", tc.step, tc.role)
			if !ok {
				t.Fatal("a saved workflow must always yield a derivation")
			}
			if got != tc.want {
				t.Fatalf("ExpectedHandoff(%s, %s) = %q, want %q", tc.step, tc.role, got, tc.want)
			}
		})
	}
}
