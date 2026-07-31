// Acceptance: specs/truthful-validators.feature
//
// Section F (part 3) — gettext is unaffected by the single-locale warning, an
// i18n gate filtered out of the diff scope reports SKIP, and a run with only
// SKIP/WARN gates still passes overall.
package acceptance_test

import "testing"

// Scenario: The gettext path is unaffected by the single-locale warning
func TestTV_G11_GettextSingleLocaleUnaffected(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupGitRepoWithCleanBranch(t)
	writeFile(t, dir, "centinela.toml",
		"[gates]\ni18n = true\n\n[i18n]\nformat = \"gettext\"\ndir = \"locales\"\nlocales = [\"en\"]\n")
	writeFile(t, dir, "locales/en.po", "msgid \"x\"\nmsgstr \"ok\"\n")
	commit(t, dir, "add gettext locale")
	out := runValidate(t, bin, dir, nil)
	mustNotContain(t, out, "trivially satisfied")
}

// Scenario: An i18n gate filtered out of the diff scope reports skipped
//
// The locale files must sit on the MAIN baseline, not the feature diff — a
// diff that itself touched a locale file would legitimately run the gate.
func TestTV_G11_FilteredOutOfDiffScopeReportsSkip(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@centinela.dev")
	runGit(t, dir, "config", "user.name", "Test")
	writeFile(t, dir, "centinela.toml",
		"[gates]\ni18n = true\n\n[i18n]\nformat = \"json\"\ndir = \"locales\"\nlocales = [\"en\", \"es\"]\n")
	writeFile(t, dir, "locales/en.json", `{"a":"x"}`)
	writeFile(t, dir, "locales/es.json", `{"a":"x"}`)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "baseline with synced locales")
	runGit(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "src/unrelated.go", "package x\n")
	commit(t, dir, "unrelated change")
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "G11")
	mustContain(t, out, "nothing inspected")
}

// Scenario: A skipped gate never turns a green run red
func TestTV_G11_SkippedGateNeverTurnsGreenRunRed(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := setupGitRepoWithCleanBranch(t)
	out := runValidate(t, bin, dir, []string{"--changed"})
	mustContain(t, out, "All gates passed")
	mustContain(t, out, "— G1: File Size")
}
