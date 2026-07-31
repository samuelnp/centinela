// Acceptance: specs/truthful-validators.feature
//
// Section F (part 1) — G1 (file size) reports what it actually inspected: an
// empty diff scope is SKIP, a clean scoped pass names the cap, the cap and
// Fail severity are unchanged, and a justified exception still passes with
// its own message. Fixtures reuse the git helpers from
// diff_aware_gatekeeper_acceptance_test.go (same package).
package acceptance_test

import "testing"

// Scenario: A file-size gate that inspected nothing reports skipped
func TestTV_G1_EmptyDiffScopeReportsSkip(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupGitRepoWithCleanBranch(t)
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "G1: File Size")
	mustContain(t, out, "nothing inspected")
	mustNotContain(t, out, "within the")
}

// Scenario: A clean file-size pass names the cap and the exception mechanism
func TestTV_G1_CleanScopedPassNamesCapAndExceptions(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupGitRepoWithCleanBranch(t)
	writeFile(t, dir, "src/small.go", "package x\n")
	commit(t, dir, "add small file")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "G1: File Size")
	mustContain(t, out, "100-line cap")
	mustContain(t, out, "exceptions are configurable")
}

// Scenario: The file-size cap and severity are unchanged
func TestTV_G1_CapAndSeverityUnchanged(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupGitRepoWithCleanBranch(t)
	writeOversized(t, dir, "src/new.go")
	commit(t, dir, "add oversized")
	out := runValidateExpectFail(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "src/new.go")
	mustContain(t, out, "100")
}

// Scenario: A justified oversized file still passes with its own message
//
// RenderGateResult only prints Details on a Fail (a Pass line stays one-line
// muted text), so the file/line-count/justification content the spec asks for
// lives in the Result value, not the rendered CLI text — that is asserted
// directly against checkFileSize in
// internal/gates/file_size_truthful_test.go's
// TestCheckFileSize_JustifiedPassKeepsItsOwnMessage. This test pins the
// CLI-visible half: the run still exits 0 and names the justified-exception
// pass distinctly from an unqualified clean pass.
func TestTV_G1_JustifiedExceptionStillPasses(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupGitRepoWithCleanBranch(t)
	writeOversized(t, dir, "src/big.go") // 101 lines, within the 130 exception cap
	writeFile(t, dir, "centinela.toml",
		"[[gates.file_size_exceptions]]\n"+
			"path = \"src/big.go\"\n"+
			"kind = \"domain_atomic\"\n"+
			"reason = \"generated fixture, split would fragment one table\"\n"+
			"max_lines = 130\n")
	commit(t, dir, "add justified big file")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "justified exception")
}
