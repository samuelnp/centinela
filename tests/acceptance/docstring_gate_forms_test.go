package acceptance_test

import "testing"

// Acceptance: specs/docstring-gate.feature
//
// What counts as documented: a block doc on a grouped const covers every
// spec inside it, an exempted identifier passes and stays visible in the
// report, undocumented struct fields do not fail the default configuration,
// and a file that fails to parse is reported rather than silently dropped.

// Scenario: A grouped const block doc covers every constant in the block
func TestDG_GroupedConstBlockDocCovers(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "a.go", "package a\n\n"+
		"// Values used by this package.\n"+
		"const (\n\tA = 1\n\tB = 2\n)\n")
	commit(t, dir, "add grouped const block")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "docstring-gate")
	mustContain(t, out, "All 2 exported identifiers across 1 changed Go file(s) are documented")
}

// Scenario: A nodoc directive exempts an identifier and the exemption is reported
//
// The invisibility bug: RenderGateResult hides Details on Pass, so this
// must be asserted through `docs lint`, which echoes them back.
func TestDG_NodocExemptionListedOnPass(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "a.go", "package a\n\n//centinela:nodoc\nfunc Exported() {}\n")
	commit(t, dir, "add exempt export")
	out, code := runCent(t, bin, dir, "docs", "lint")
	if code != 0 {
		t.Fatalf("an exempted-only file must pass: %d\n%s", code, out)
	}
	mustContain(t, out, "src/a.go:4: func Exported exempt via //centinela:nodoc")
}

// Scenario: Struct fields are not reported when check_fields is false
func TestDG_StructFieldsNotReportedByDefault(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "a.go", "package a\n\n"+
		"// T is documented.\ntype T struct {\n\tF string\n}\n")
	commit(t, dir, "add struct with undocumented field")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "docstring-gate")
	mustContain(t, out, "All 1 exported identifiers across 1 changed Go file(s) are documented")
}

// Scenario: An unparseable Go file is reported, not silently skipped
func TestDG_UnparseableFileReported(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepo(t, "fail")
	writeDocFile(t, dir, "broken.go", "package a\n\nfunc (((\n")
	commit(t, dir, "add broken file")
	out := runValidateExpectFail(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "docstring-gate")
	mustContain(t, out, "src/broken.go: unparseable Go file")
}
