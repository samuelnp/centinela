package gates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

func TestCheckFileSize_FilterEmptySetSkipsAllFiles(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src", 0755) //nolint:errcheck
	big := makeBigSource(101)
	os.WriteFile("src/big.go", []byte(big), 0644) //nolint:errcheck

	r := checkFileSize(&config.Config{}, gitdiff.NewSet(nil))
	if r.Status != Pass {
		t.Fatalf("empty filter must pass, got %v: %v", r.Status, r.Details)
	}
	if r.Message != "No relevant changes — gate skipped." {
		t.Fatalf("expected skip message, got %q", r.Message)
	}
}

func TestCheckFileSize_FilterRestrictsScannedSet(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src", 0755) //nolint:errcheck
	big := makeBigSource(101)
	os.WriteFile("src/included.go", []byte(big), 0644) //nolint:errcheck
	os.WriteFile("src/ignored.go", []byte(big), 0644)  //nolint:errcheck

	r := checkFileSize(&config.Config{}, gitdiff.NewSet([]string{"src/included.go"}))
	if r.Status != Fail {
		t.Fatalf("expected fail for included file, got %v", r.Status)
	}
	if len(r.Details) != 1 || !contains(r.Details[0], "included.go") {
		t.Fatalf("expected only included.go in details, got %v", r.Details)
	}
	for _, d := range r.Details {
		if contains(d, "ignored.go") {
			t.Fatalf("ignored.go must not be flagged when outside filter")
		}
	}
}

func TestCheckI18nFiltered_SkipsWhenNoLocaleChanged(t *testing.T) {
	cfg := &config.Config{
		Gates: config.GatesConfig{I18nEnabled: true},
		I18n:  config.I18nConfig{Format: "json", Dir: "src/i18n", Locales: []string{"en"}},
	}
	r := checkI18nFiltered(cfg, gitdiff.NewSet([]string{"cmd/main.go"}))
	if r.Status != Pass || r.Message != "No locale changes — gate skipped." {
		t.Fatalf("expected skip pass, got %+v", r)
	}
}

func TestCheckI18nFiltered_RunsWhenLocaleChanged(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src/i18n", 0755)                   //nolint:errcheck
	writeJSON(t, "src/i18n/en.json", map[string]any{"hello": "Hello"})
	writeJSON(t, "src/i18n/es.json", map[string]any{"hello": "Hola"})
	cfg := &config.Config{
		Gates: config.GatesConfig{I18nEnabled: true},
		I18n:  config.I18nConfig{Format: "json", Dir: "src/i18n", Locales: []string{"en", "es"}},
	}
	r := checkI18nFiltered(cfg, gitdiff.NewSet([]string{"src/i18n/es.json"}))
	if r.Status != Pass || r.Message != "All locales have identical keys." {
		t.Fatalf("expected full G11 pass, got %+v", r)
	}
}

func TestCheckI18nFiltered_NilFilterFallsThroughToFullCheck(t *testing.T) {
	cfg := &config.Config{
		Gates: config.GatesConfig{I18nEnabled: true},
		I18n:  config.I18nConfig{Format: "", Locales: nil},
	}
	r := checkI18nFiltered(cfg, nil)
	if r.Status != Skip {
		t.Fatalf("nil filter must defer to checkI18n; expected Skip, got %+v", r)
	}
}

func makeBigSource(lines int) string {
	out := ""
	for i := 0; i < lines; i++ {
		out += "x\n"
	}
	return out
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Clean(path), data, 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
