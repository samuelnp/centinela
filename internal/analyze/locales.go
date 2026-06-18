package analyze

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// localeDirs are conventional i18n directory names searched at the repo root and
// one level below (e.g. src/locales). Their immediate entries are read as locale
// codes (file stem or subdirectory name).
var localeDirs = []string{"locales", "locale", "i18n", "lang", "translations"}

// localeCode matches a bare locale code: a two-letter language, optionally with
// a region suffix (e.g. "en", "es", "pt-BR", "zh_CN").
var localeCode = regexp.MustCompile(`^[a-z]{2}([-_][A-Za-z]{2,4})?$`)

// detectLocales returns the sorted, de-duplicated set of locale codes found in
// known i18n directories under root. No i18n present yields an empty slice (not
// an error).
func detectLocales(root string) []string {
	found := map[string]bool{}
	for _, base := range candidateLocaleDirs(root) {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if !e.IsDir() {
				name = strings.TrimSuffix(name, filepath.Ext(name))
			}
			if localeCode.MatchString(name) {
				found[name] = true
			}
		}
	}
	out := make([]string, 0, len(found))
	for code := range found {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

// candidateLocaleDirs returns existing locale directories at the root and under
// a single intermediate directory (e.g. src/locales).
func candidateLocaleDirs(root string) []string {
	var dirs []string
	for _, name := range localeDirs {
		dirs = append(dirs, filepath.Join(root, name))
	}
	if subs, err := os.ReadDir(root); err == nil {
		for _, s := range subs {
			if s.IsDir() && !skipDirs[s.Name()] {
				for _, name := range localeDirs {
					dirs = append(dirs, filepath.Join(root, s.Name(), name))
				}
			}
		}
	}
	return dirs
}
