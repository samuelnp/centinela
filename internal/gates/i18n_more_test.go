package gates

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestCheckI18nGettextMissingFile(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("i18n", 0755)                                                //nolint:errcheck
	os.WriteFile("i18n/en.po", []byte("msgid \"x\"\nmsgstr \"ok\"\n"), 0644) //nolint:errcheck
	r := checkI18nGettext(config.I18nConfig{Dir: "i18n", Locales: []string{"en", "es"}})
	if r.Status != Fail {
		t.Fatalf("expected fail on missing gettext locale, got %v", r.Status)
	}
}
