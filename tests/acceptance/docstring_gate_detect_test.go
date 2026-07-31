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
	mustContain(t, out, "docstring-gate  1 undocumented exported identifier")
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
// RenderGateResult only expands Details on Fail, so a Warn must carry its own
// count and point at the surface that can list them — it must never end in a
// colon introducing a list nobody will see. The per-violation identifier is
// pinned on the gate surface by TestReportDocstring_* and via `docs lint` in
// TestDG_DocsLintExitCodes.
func TestDG_WarnSeverityDoesNotFail(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "warn")
	writeDocFile(t, dir, "a.go", "package a\n\nfunc Exported() {}\n")
	commit(t, dir, "add undocumented export")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "⚠ docstring-gate  1 undocumented exported identifier in changed files")
	mustContain(t, out, "centinela docs lint")
	mustNotContain(t, out, "in changed files —\n")
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
