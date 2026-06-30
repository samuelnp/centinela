package gates

import (
	"testing"

	"github.com/samuelnp/centinela/internal/gitdiff"
)

// TestLocaleFileInFilter_EmptyDirAlwaysTrue verifies that an unconfigured i18n
// dir is treated as "always in scope" so the gate never silently skips.
func TestLocaleFileInFilter_EmptyDirAlwaysTrue(t *testing.T) {
	if !localeFileInFilter("", gitdiff.NewSet([]string{"cmd/main.go"})) {
		t.Fatal("empty dir must report the locale as in-filter")
	}
}

// TestLocaleFileInFilter_TrailingSlashNormalized verifies the configured dir is
// matched whether or not it carries a trailing slash.
func TestLocaleFileInFilter_TrailingSlashNormalized(t *testing.T) {
	f := gitdiff.NewSet([]string{"src/i18n/en.json"})
	if !localeFileInFilter("src/i18n/", f) {
		t.Fatal("trailing-slash dir must match a file under it")
	}
	if !localeFileInFilter("src/i18n", f) {
		t.Fatal("dir without trailing slash must match a file under it")
	}
	if localeFileInFilter("src/i18n", gitdiff.NewSet([]string{"cmd/main.go"})) {
		t.Fatal("unrelated change must not be reported as in-filter")
	}
}
