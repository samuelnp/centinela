package gates

import (
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// checkI18nFiltered short-circuits the i18n gate when no locale file is in
// the diff filter. Key-completeness can only be compared whole-repo, so a
// partial scan is not meaningful; gating on "any locale file touched" is
// the diff-aware contract for G11.
func checkI18nFiltered(cfg *config.Config, filter *gitdiff.Set) Result {
	if filter != nil && !localeFileInFilter(cfg.I18n.Dir, filter) {
		return Result{
			Name:    "G11: i18n",
			Status:  Pass,
			Message: "No locale changes — gate skipped.",
		}
	}
	return checkI18n(cfg)
}

func localeFileInFilter(dir string, filter *gitdiff.Set) bool {
	if dir == "" {
		return true
	}
	prefix := strings.TrimSuffix(dir, "/") + "/"
	return filter.HasPrefix(prefix)
}
