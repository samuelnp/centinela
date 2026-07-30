package gates

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// i18nFixture chdirs into a temp repo with an i18n dir and returns nothing —
// the caller writes whichever locale files the case needs.
func i18nFixture(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("i18n", 0o755); err != nil {
		t.Fatal(err)
	}
}

// One locale is not a parity result: the file is still read and parsed, but the
// gate says the check verified nothing instead of claiming identical keys.
func TestCheckI18nJSON_SingleLocaleWarns(t *testing.T) {
	i18nFixture(t)
	if err := os.WriteFile("i18n/en.json", []byte(`{"a":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	r := checkI18nJSON(config.I18nConfig{Dir: "i18n", Locales: []string{"en"}})
	if r.Status != Warn {
		t.Fatalf("one locale must Warn, got %+v", r)
	}
	if !strings.Contains(r.Message, "trivially satisfied") {
		t.Fatalf("message must state the check is trivial, got %q", r.Message)
	}
	if strings.Contains(r.Message, "identical keys") {
		t.Fatalf("message must not claim all locales have identical keys: %q", r.Message)
	}
	// R6: single-locale repos must be told this Warn is advisory and where it
	// could still bite (pr_gate.fail_on_warning, off by default).
	if !strings.Contains(r.Message, "fail_on_warning") {
		t.Fatalf("message must name the fail_on_warning interaction, got %q", r.Message)
	}
}

// The Fail paths above the warning are unreachable-by-construction, not skipped:
// a missing or malformed single locale file still Fails.
func TestCheckI18nJSON_SingleLocaleStillFailsOnBadFile(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		i18nFixture(t)
		r := checkI18nJSON(config.I18nConfig{Dir: "i18n", Locales: []string{"en"}})
		if r.Status != Fail {
			t.Fatalf("a missing single locale file must Fail, got %+v", r)
		}
	})
	t.Run("malformed", func(t *testing.T) {
		i18nFixture(t)
		if err := os.WriteFile("i18n/en.json", []byte(`{"a":`), 0o644); err != nil {
			t.Fatal(err)
		}
		r := checkI18nJSON(config.I18nConfig{Dir: "i18n", Locales: []string{"en"}})
		if r.Status != Fail {
			t.Fatalf("a malformed single locale file must Fail, got %+v", r)
		}
	})
}

// Two locales are unchanged in both directions.
func TestCheckI18nJSON_TwoLocalesUnchanged(t *testing.T) {
	i18nFixture(t)
	if err := os.WriteFile("i18n/en.json", []byte(`{"a":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("i18n/es.json", []byte(`{"a":"y"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.I18nConfig{Dir: "i18n", Locales: []string{"en", "es"}}
	if r := checkI18nJSON(cfg); r.Status != Pass {
		t.Fatalf("two locales in sync must still Pass, got %+v", r)
	}
	if err := os.WriteFile("i18n/es.json", []byte(`{"b":"y"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := checkI18nJSON(cfg); r.Status != Fail {
		t.Fatalf("two locales out of sync must still Fail, got %+v", r)
	}
}

// gettext is untouched: per-locale msgstr completeness is meaningful with one
// locale, so it stays a real Pass and never emits the parity warning.
func TestCheckI18nGettext_SingleLocaleStillPasses(t *testing.T) {
	i18nFixture(t)
	if err := os.WriteFile("i18n/en.po", []byte("msgid \"x\"\nmsgstr \"ok\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := checkI18nGettext(config.I18nConfig{Dir: "i18n", Locales: []string{"en"}})
	if r.Status != Pass || strings.Contains(r.Message, "trivially satisfied") {
		t.Fatalf("gettext single locale must stay a real pass, got %+v", r)
	}
}
