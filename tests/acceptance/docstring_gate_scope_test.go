package acceptance_test

import (
	"path/filepath"
	"testing"
)

// Acceptance: specs/docstring-gate.feature
//
// What the gate accepts as documentation, and what it refuses to treat as
// source at all. Driven end-to-end through the real binary because both
// defects shipped as gate-only behavior that `docs lint --full` did not share.
// runValidate sets CI= in the child env, so the run resolves diff-aware mode.

// Scenario: A trailing line comment counts as documentation
func TestDG_TrailingLineCommentIsDocumentation(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "forms.go", "package a\n\n"+
		"const Trailing = 1 // Trailing is the answer.\n\n"+
		"var VarTrailing = 2 // VarTrailing is a var.\n")
	commit(t, dir, "add trailing-comment forms")
	out := runValidate(t, bin, dir, nil)
	mustContain(t, out, "All 2 exported identifiers")
	mustNotContain(t, out, "has no doc comment")
	mustContain(t, out, "All gates passed")
}

// excludedFixtures are files the gate must refuse to treat as source.
var excludedFixtures = map[string]string{
	"vendor/v.go":       "package v\n\nfunc VendoredExport() {}\n",
	"testdata/bad.go":   "package t\n\nfunc (((\n",
	"node_modules/n.go": "package n\n\nfunc NodeExport() {}\n",
	"dist/d.go":         "package d\n\nfunc DistExport() {}\n",
	".worktrees/f/w.go": "package w\n\nfunc WorktreeExport() {}\n",
}

// Scenario: Vendored, build and fixture directories are never inspected
func TestDG_ExcludedDirectoriesAreNeverInspected(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "ok.go", "package a\n\n// Ok is documented.\nfunc Ok() {}\n")
	for rel, body := range excludedFixtures {
		mustWrite(t, filepath.Join(dir, "src", rel), body)
	}
	commit(t, dir, "add excluded-directory files")

	out := runValidate(t, bin, dir, nil)
	mustContain(t, out, "All 1 exported identifiers")
	mustContain(t, out, "All gates passed")
	for _, unwanted := range []string{
		"VendoredExport", "NodeExport", "DistExport", "WorktreeExport", "unparseable",
	} {
		mustNotContain(t, out, unwanted)
	}

	// The report surface must agree with the gate on the identical tree —
	// otherwise --full cannot be used to predict or audit what the gate does.
	full, code := runCent(t, bin, dir, "docs", "lint", "--full")
	if code != 0 {
		t.Fatalf("docs lint --full must stay report-only, exit %d:\n%s", code, full)
	}
	mustContain(t, full, "0 undocumented of 1 exported identifiers")
}
