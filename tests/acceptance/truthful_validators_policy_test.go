// Acceptance: specs/truthful-validators.feature
//
// Section C — [validate] acceptance_skip_policy is configurable and defaults
// to fail. An unknown value is a config error, never silently normalized.
package acceptance_test

import "testing"

const tvSkipCmd = `printf '3 scenarios (1 skipped, 2 passed)\n' # tests/acceptance`

// Scenario: An absent policy key behaves as fail
func TestTV_Policy_AbsentKeyBehavesAsFail(t *testing.T) {
	out, code := tvValidateWithCmd(t, tvSkipCmd, "")
	if code == 0 {
		t.Fatalf("no acceptance_skip_policy key must default to fail\n%s", out)
	}
}

// Scenario: The warn policy surfaces the skip without failing the run
func TestTV_Policy_WarnSurfacesWithoutFailing(t *testing.T) {
	out, code := tvValidateWithCmd(t, tvSkipCmd, "warn")
	if code != 0 {
		t.Fatalf("warn policy must exit zero\n%s", out)
	}
	mustContain(t, out, "1 skipped")
}

// Scenario: The off policy restores the previous exit-code-only behavior
func TestTV_Policy_OffRestoresExitCodeOnlyBehavior(t *testing.T) {
	out, code := tvValidateWithCmd(t, tvSkipCmd, "off")
	if code != 0 {
		t.Fatalf("off policy must pass regardless of reported skips\n%s", out)
	}
}

// Scenario: An unknown policy value is a configuration error
func TestTV_Policy_UnknownValueIsAConfigError(t *testing.T) {
	out, code := tvValidateWithCmd(t, tvSkipCmd, "maybe")
	if code == 0 {
		t.Fatalf("an unknown policy value must fail config load\n%s", out)
	}
	mustContain(t, out, "fail")
	mustContain(t, out, "warn")
	mustContain(t, out, "off")
	if code == 0 {
		t.Fatal("the value must not be silently normalized to the default")
	}
}
