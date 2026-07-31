package acceptance_test

import "testing"

// Acceptance: specs/docstring-gate.feature
//
// Core detection: an undocumented export in a changed file fails at "fail"
// severity and warns-without-blocking at "warn", a fully documented change
// passes naming the inspected count, and a legacy undocumented file outside
// the changed-file scope is never opened.

// Scenario: Undocumented exported function in a changed file fails the gate
func TestDG_UndocumentedExportFails(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "a.go", "package a\n\nfunc Exported() {}\n")
	commit(t, dir, "add undocumented export")
	out := runValidateExpectFail(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "docstring-gate  Undocumented exported identifiers")
	mustContain(t, out, "src/a.go:3: func Exported has no doc comment")
}

// Scenario: Documented exported identifiers pass
func TestDG_DocumentedPasses(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "a.go", "package a\n\n// Exported does a thing.\nfunc Exported() {}\n")
	commit(t, dir, "add documented export")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "docstring-gate")
	mustContain(t, out, "All 1 exported identifiers across 1 changed Go file(s) are documented")
}

// Scenario: Warn severity reports the violations without failing
//
// RenderGateResult only expands Details on Fail (Warn stays a one-line
// summary here), so the per-violation identifier is pinned separately via
// `docs lint` in TestDG_DocsLintExitCodes; this test pins the CLI-visible
// half: the run stays Warn and gates.AllPassed keeps validate exiting 0.
func TestDG_WarnSeverityDoesNotFail(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "warn")
	writeDocFile(t, dir, "a.go", "package a\n\nfunc Exported() {}\n")
	commit(t, dir, "add undocumented export")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "⚠ docstring-gate  Undocumented exported identifiers in changed files:")
	mustContain(t, out, "All gates passed")
}

// Scenario: A legacy unchanged file with undocumented exports is never scanned
func TestDG_LegacyUnchangedNeverScanned(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepoWithLegacy(t, "fail", "package a\n\nfunc Legacy() {}\nfunc AlsoLegacy() {}\n")
	writeDocFile(t, dir, "new.go", "package a\n\n// New is documented.\nfunc New() {}\n")
	commit(t, dir, "add documented change")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "docstring-gate")
	mustContain(t, out, "All 1 exported identifiers across 1 changed Go file(s) are documented")
	mustNotContain(t, out, "legacy.go")
}
