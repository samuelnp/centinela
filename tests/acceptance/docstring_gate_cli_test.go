package acceptance_test

import (
	"path/filepath"
	"testing"
)

// Acceptance: specs/docstring-gate.feature
//
// The `centinela docs lint` CLI surface: config validation, fail/warn exit
// codes, the --full backlog report, and the --json shape.

// Scenario: An unknown severity is a configuration error
func TestDG_UnknownSeverityConfigError(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "centinela.toml"),
		"[gates.docstring]\nenabled = true\nseverity = \"nope\"\n")
	out, code := runCent(t, bin, dir, "docs", "lint")
	if code == 0 {
		t.Fatalf("unknown severity must be a config error:\n%s", out)
	}
	mustContain(t, out, "gates.docstring.severity")
	mustContain(t, out, "fail")
	mustContain(t, out, "warn")
}

// Scenario: centinela docs lint exits 1 on Fail and 0 on Warn
func TestDG_DocsLintExitCodes(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dirFail := setupDocstringRepo(t, "fail")
	writeDocFile(t, dirFail, "a.go", "package a\n\nfunc Exported() {}\n")
	commit(t, dirFail, "add undocumented export")
	outFail, codeFail := runCent(t, bin, dirFail, "docs", "lint")
	if codeFail == 0 {
		t.Fatalf("fail severity must exit non-zero:\n%s", outFail)
	}
	mustContain(t, outFail, "Exported")

	dirWarn := setupDocstringRepo(t, "warn")
	writeDocFile(t, dirWarn, "a.go", "package a\n\nfunc Exported() {}\n")
	commit(t, dirWarn, "add undocumented export")
	outWarn, codeWarn := runCent(t, bin, dirWarn, "docs", "lint")
	if codeWarn != 0 {
		t.Fatalf("warn severity must exit 0:\n%s", outWarn)
	}
	mustContain(t, outWarn, "Exported")
}

// Scenario: centinela docs lint --full reports the legacy backlog without failing
func TestDG_DocsLintFullBacklog(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepoWithLegacy(t, "fail", "package a\n\nfunc Legacy() {}\n")
	out, code := runCent(t, bin, dir, "docs", "lint", "--full")
	if code != 0 {
		t.Fatalf("--full must always exit 0: %d\n%s", code, out)
	}
	mustContain(t, out, "1 undocumented of 1 exported identifiers")
	mustContain(t, out, "src/legacy.go:3: func Legacy")
}

// docs lint --json shape (required end-to-end coverage; not a separate
// Gherkin scenario, extends "docs lint exits 1 on Fail and 0 on Warn").
func TestDG_DocsLintJSONShape(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "a.go", "package a\n\nfunc Exported() {}\n")
	commit(t, dir, "add undocumented export")
	out, code := runCent(t, bin, dir, "docs", "lint", "--json")
	if code == 0 {
		t.Fatalf("json mode must still exit non-zero on fail:\n%s", out)
	}
	for _, want := range []string{`"scope": "changed"`, `"status": "fail"`,
		`"undocumented": 1`, "src/a.go:3: func Exported"} {
		mustContain(t, out, want)
	}
}
