package synthesize

import (
	"strings"

	"github.com/samuelnp/centinela/internal/analyze"
)

// signals is the lowercased, flattened view of an Inventory the rule predicates
// match against, precomputed once per inference so rules stay cheap.
type signals struct {
	lang     string
	pkgs     []string // lowercased package paths
	deps     []string // lowercased dependency names (all manifests)
	frames   []string // lowercased framework strings (all manifests)
	kinds    []string // manifest kinds (e.g. "gem", "go-mod")
	hasGraph bool
}

func newSignals(inv analyze.Inventory) signals {
	s := signals{lang: strings.ToLower(inv.PrimaryLanguage), hasGraph: len(inv.Graph.Edges) > 0}
	for _, p := range inv.Packages {
		s.pkgs = append(s.pkgs, strings.ToLower(p))
	}
	for _, m := range inv.Manifests {
		s.kinds = append(s.kinds, m.Kind)
		if m.Framework != "" {
			s.frames = append(s.frames, strings.ToLower(m.Framework))
		}
		for _, d := range m.Deps {
			s.deps = append(s.deps, strings.ToLower(d))
		}
	}
	return s
}

// hasPkg reports whether any package path contains sub.
func (s signals) hasPkg(sub string) bool { return anyContains(s.pkgs, sub) }

// hasDep reports whether any dependency name contains sub.
func (s signals) hasDep(sub string) bool { return anyContains(s.deps, sub) }

// hasFramework reports whether any framework string contains sub.
func (s signals) hasFramework(sub string) bool { return anyContains(s.frames, sub) }

// hasKind reports whether a manifest of the given kind was detected.
func (s signals) hasKind(kind string) bool {
	for _, k := range s.kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func anyContains(hay []string, sub string) bool {
	for _, h := range hay {
		if strings.Contains(h, sub) {
			return true
		}
	}
	return false
}
