package gates

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestAllPassed(t *testing.T) {
	if !AllPassed([]Result{{Status: Pass}, {Status: Warn}}) {
		t.Fatal("expected pass when no fail results")
	}
	if AllPassed([]Result{{Status: Fail}}) {
		t.Fatal("expected false when fail exists")
	}
}

func TestRunAllIncludesEnabledGates(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src/i18n", 0755)                                   //nolint:errcheck
	os.WriteFile("src/i18n/en.json", []byte(`{"hello":"x"}`), 0644) //nolint:errcheck
	os.WriteFile("src/i18n/es.json", []byte(`{"hello":"y"}`), 0644) //nolint:errcheck
	cfg := &config.Config{Gates: config.GatesConfig{FileSizeEnabled: true, I18nEnabled: true}, I18n: config.I18nConfig{Format: "json", Dir: "src/i18n", Locales: []string{"en", "es"}}}
	res := RunAll(cfg)
	if len(res) != 2 {
		t.Fatalf("expected 2 gate results, got %d", len(res))
	}
}
