package gates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

func checkI18n(cfg *config.Config) Result {
	i := cfg.I18n
	if i.Format == "" || i.Format == "none" {
		return Result{Name: "G11: i18n", Status: Skip, Message: "i18n gate skipped (no format configured)."}
	}
	if len(i.Locales) == 0 {
		return Result{Name: "G11: i18n", Status: Skip, Message: "i18n gate skipped (no locales configured)."}
	}

	switch i.Format {
	case "json":
		return checkI18nJSON(i)
	case "gettext":
		return checkI18nGettext(i)
	default:
		return Result{
			Name:    "G11: i18n",
			Status:  Warn,
			Message: fmt.Sprintf("Unknown i18n format %q — skipping built-in check.", i.Format),
		}
	}
}

// checkI18nJSON verifies all locales have the same keys as the first locale.
func checkI18nJSON(i config.I18nConfig) Result {
	keys := map[string]map[string]bool{} // locale → key set

	for _, locale := range i.Locales {
		path := filepath.Join(i.Dir, locale+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			return Result{
				Name:    "G11: i18n",
				Status:  Fail,
				Message: "Locale file missing.",
				Details: []string{filepath.ToSlash(path)},
			}
		}

		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return Result{
				Name:    "G11: i18n",
				Status:  Fail,
				Message: fmt.Sprintf("Cannot parse %s: %s", path, err),
			}
		}

		keys[locale] = flatKeys(m, "")
	}

	return compareKeysets(keys, i.Locales)
}

// flatKeys flattens a nested JSON map to dot-separated keys.
func flatKeys(m map[string]interface{}, prefix string) map[string]bool {
	out := map[string]bool{}
	for k, v := range m {
		full := k
		if prefix != "" {
			full = prefix + "." + k
		}
		if nested, ok := v.(map[string]interface{}); ok {
			for sub := range flatKeys(nested, full) {
				out[sub] = true
			}
		} else {
			out[full] = true
		}
	}
	return out
}

func compareKeysets(keys map[string]map[string]bool, locales []string) Result {
	ref := locales[0]
	var missing []string

	for _, locale := range locales[1:] {
		for k := range keys[ref] {
			if !keys[locale][k] {
				missing = append(missing, fmt.Sprintf("[%s] missing key: %s", locale, k))
			}
		}
		for k := range keys[locale] {
			if !keys[ref][k] {
				missing = append(missing, fmt.Sprintf("[%s] extra key not in %s: %s", locale, ref, k))
			}
		}
	}

	if len(missing) == 0 {
		return Result{Name: "G11: i18n", Status: Pass, Message: "All locales have identical keys."}
	}
	return Result{Name: "G11: i18n", Status: Fail, Message: "Translation keys out of sync.", Details: missing}
}

// checkI18nGettext checks .po files for untranslated msgstr entries.
func checkI18nGettext(i config.I18nConfig) Result {
	var violations []string

	for _, locale := range i.Locales {
		path := filepath.Join(i.Dir, locale+".po")
		data, err := os.ReadFile(path)
		if err != nil {
			violations = append(violations, fmt.Sprintf("missing: %s", filepath.ToSlash(path)))
			continue
		}

		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == `msgstr ""` {
				violations = append(violations, fmt.Sprintf("[%s] untranslated entry in %s", locale, filepath.ToSlash(path)))
				break
			}
		}
	}

	if len(violations) == 0 {
		return Result{Name: "G11: i18n", Status: Pass, Message: "All gettext locales fully translated."}
	}
	return Result{Name: "G11: i18n", Status: Fail, Message: "Untranslated strings found.", Details: violations}
}
