package analyze

import (
	"os"
	"path/filepath"
	"strings"
)

// gitignore is a deliberately simple, dependency-free matcher over the repo's
// root .gitignore. It supports plain path/name patterns and a trailing-slash
// directory form — enough to honor the skip intent (AC-5) without pulling in a
// full gitignore engine. Negation and nested ignore files are out of scope.
type gitignore struct {
	patterns []string
}

// loadGitignore reads root/.gitignore, returning an empty matcher when absent
// or unreadable (best-effort: a missing ignore file is not an error).
func loadGitignore(root string) gitignore {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return gitignore{}
	}
	var pats []string
	for _, line := range strings.Split(string(data), "\n") {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") || strings.HasPrefix(s, "!") {
			continue
		}
		pats = append(pats, strings.TrimSuffix(strings.TrimPrefix(s, "/"), "/"))
	}
	return gitignore{patterns: pats}
}

// match reports whether a module-relative slash path is ignored. A pattern
// matches the full relative path, its basename, or any leading path segment so
// an ignored directory hides everything beneath it.
func (g gitignore) match(rel string) bool {
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)
	for _, p := range g.patterns {
		if p == rel || p == base {
			return true
		}
		if strings.HasPrefix(rel, p+"/") {
			return true
		}
	}
	return false
}
