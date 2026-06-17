package gates

import (
	"errors"
	"strings"
)

// errEmptyModule is the config error for an explicitly-blank module path.
var errEmptyModule = errors.New("module path is empty")

// trimDoubleStar returns (dir, true) for a glob ending in "/**", where dir is
// the glob with the "/**" suffix removed. A bare "**" maps to "" (matches the
// whole tree). Returns ("", false) when the glob has no "/**" suffix.
func trimDoubleStar(glob string) (string, bool) {
	if glob == "**" {
		return "", true
	}
	if strings.HasSuffix(glob, "/**") {
		return strings.TrimSuffix(glob, "/**"), true
	}
	return "", false
}

// hasPrefixDir reports whether rel is inside the directory base, i.e. base is
// a path-segment prefix of rel ("internal/config" matches "internal/config/x"
// but not "internal/configx"). An empty base matches everything.
func hasPrefixDir(rel, base string) bool {
	if base == "" {
		return true
	}
	return strings.HasPrefix(rel, base+"/")
}

// allowed reports whether a package in layer "from" may import a package in
// layer "to". Same-layer (and self) imports are always allowed.
func (m matrix) allowed(from, to string) bool {
	if from == to {
		return true
	}
	return m.allow[from][to]
}
