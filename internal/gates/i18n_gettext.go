package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

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
