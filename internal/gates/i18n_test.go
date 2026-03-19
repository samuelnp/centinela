package gates

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestCheckI18nBranches(t *testing.T) {
	if r := checkI18n(&config.Config{}); r.Status != Skip {
		t.Fatalf("expected skip for empty config, got %v", r.Status)
	}
	if r := checkI18n(&config.Config{I18n: config.I18nConfig{Format: "json"}}); r.Status != Skip {
		t.Fatalf("expected skip for no locales, got %v", r.Status)
	}
	if r := checkI18n(&config.Config{I18n: config.I18nConfig{Format: "weird", Locales: []string{"en"}}}); r.Status != Warn {
		t.Fatalf("expected warn for unknown format, got %v", r.Status)
	}
}

func TestCheckI18nJSONAndGettext(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("i18n", 0755)                                     //nolint:errcheck
	os.WriteFile("i18n/en.json", []byte(`{"a":{"b":"x"}}`), 0644) //nolint:errcheck
	os.WriteFile("i18n/es.json", []byte(`{"a":{"b":"y"}}`), 0644) //nolint:errcheck
	if r := checkI18nJSON(config.I18nConfig{Dir: "i18n", Locales: []string{"en", "es"}}); r.Status != Pass {
		t.Fatalf("json check should pass, got %v", r.Status)
	}
	os.WriteFile("i18n/en.po", []byte("msgid \"x\"\nmsgstr \"ok\"\n"), 0644) //nolint:errcheck
	os.WriteFile("i18n/es.po", []byte("msgid \"x\"\nmsgstr \"\"\n"), 0644)   //nolint:errcheck
	if r := checkI18nGettext(config.I18nConfig{Dir: "i18n", Locales: []string{"en", "es"}}); r.Status != Fail {
		t.Fatalf("gettext should fail on untranslated entry, got %v", r.Status)
	}
}
