package gates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	// One locale: every file above was still read and parsed (a missing or
	// malformed file already failed), but compareKeysets' locales[1:] loop
	// would compare nothing and return a confident "identical keys" pass.
	// Say what was actually verified instead.
	if len(i.Locales) == 1 {
		return Result{
			Name:   "G11: i18n",
			Status: Warn,
			Message: "1 locale configured — the key-parity check is trivially satisfied and " +
				"verifies nothing. Advisory: it does not fail validate, but pr_gate.fail_on_warning " +
				"(off by default) would surface it.",
		}
	}
	return compareKeysets(keys, i.Locales)
}
