// Acceptance: specs/truthful-validators.feature
//
// Section F (part 2) — G11 (i18n) single-locale trivial-parity WARN, still
// Fail on a missing or malformed file, and two locales behave as before.
package acceptance_test

import "testing"

// tvI18nRepo builds a git repo with a "main" baseline and a clean feature
// branch, a centinela.toml enabling the json i18n gate for the given locales,
// and the given locale file contents (path -> content) committed on top.
func tvI18nRepo(t *testing.T, locales []string, files map[string]string) string {
	t.Helper()
	dir := setupGitRepoWithCleanBranch(t)
	toml := "[gates]\ni18n = true\n\n[i18n]\nformat = \"json\"\ndir = \"locales\"\nlocales = ["
	for i, l := range locales {
		if i > 0 {
			toml += ", "
		}
		toml += `"` + l + `"`
	}
	toml += "]\n"
	writeFile(t, dir, "centinela.toml", toml)
	for path, content := range files {
		writeFile(t, dir, path, content)
	}
	commit(t, dir, "add locale files")
	return dir
}

// Scenario: A single configured locale is reported as a trivial parity check
func TestTV_G11_SingleLocaleWarnsTrivialParity(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := tvI18nRepo(t, []string{"en"}, map[string]string{"locales/en.json": `{"a":"x"}`})
	out := runValidate(t, bin, dir, nil)
	mustContain(t, out, "trivially satisfied")
	mustNotContain(t, out, "identical keys")
}

// Scenario: A single locale with a missing file still fails
func TestTV_G11_SingleLocaleMissingFileStillFails(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := tvI18nRepo(t, []string{"en"}, map[string]string{"locales/.gitkeep": ""})
	out := runValidateExpectFail(t, bin, dir, nil)
	mustContain(t, out, "G11")
}

// Scenario: A single locale with a malformed file still fails
func TestTV_G11_SingleLocaleMalformedFileStillFails(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := tvI18nRepo(t, []string{"en"}, map[string]string{"locales/en.json": `{"a":`})
	out := runValidateExpectFail(t, bin, dir, nil)
	mustContain(t, out, "G11")
}

// Scenario: Two locales in sync still pass and out of sync still fail
func TestTV_G11_TwoLocalesSyncAndOutOfSync(t *testing.T) {
	bin := buildCentinelaBinary(t)
	dir := tvI18nRepo(t, []string{"en", "es"}, map[string]string{
		"locales/en.json": `{"a":"x"}`,
		"locales/es.json": `{"a":"y"}`,
	})
	out := runValidate(t, bin, dir, nil)
	mustContain(t, out, "G11")
	mustNotContain(t, out, "trivially satisfied")

	writeFile(t, dir, "locales/es.json", `{"b":"y"}`)
	commit(t, dir, "diverge es")
	out = runValidateExpectFail(t, bin, dir, nil)
	mustContain(t, out, "G11")
}
