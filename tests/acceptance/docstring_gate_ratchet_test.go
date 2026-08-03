package acceptance_test

import "testing"

// Acceptance: specs/docstring-gate.feature

// Scenario: The gate ratchets to changed files even when the run asks for a full scan
//
// CI mode resolves a nil filter for every gate, but the docstring gate must
// never open the legacy backlog that decision would otherwise expose. Only
// the changed, documented file may appear; the undocumented legacy file
// must stay invisible.
func TestDG_FullScanStillRatchets(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupDocstringRepoWithLegacy(t, "fail", "package a\n\nfunc Legacy() {}\n")
	writeDocFile(t, dir, "new.go", "package a\n\n// New is documented.\nfunc New() {}\n")
	commit(t, dir, "add documented change")
	out := runValidate(t, bin, dir, []string{"CI=true"})
	mustContain(t, out, "Built-in Gates (full scan)")
	mustContain(t, out, "docstring-gate")
	mustContain(t, out, "All 1 exported identifiers across 1 changed Go file(s) are documented")
	mustNotContain(t, out, "legacy.go")
	mustNotContain(t, out, "Legacy has no doc comment")
}
