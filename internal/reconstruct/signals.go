package reconstruct

import (
	"strings"

	"github.com/samuelnp/centinela/internal/analyze"
)

// signals is the lowercased, flattened view of an Inventory the selection rules
// match against, precomputed once per run so the rule table stays cheap and
// free of map-iteration order. inEdges records, per lowercased package path,
// whether any graph edge points into it (a consumed surface).
type signals struct {
	lang    string
	pkgs    []string        // lowercased package paths, original order preserved
	deps    []string        // lowercased dependency names (all manifests)
	frames  []string        // lowercased framework strings (all manifests)
	kinds   []string        // manifest kinds (e.g. "gem", "go-mod")
	inEdges map[string]bool // lowercased package path -> has an incoming graph edge
}

func newSignals(inv analyze.Inventory) signals {
	s := signals{lang: strings.ToLower(inv.PrimaryLanguage), inEdges: map[string]bool{}}
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
	for _, e := range inv.Graph.Edges {
		s.inEdges[strings.ToLower(e.To)] = true
	}
	return s
}

// hasDep reports whether any dependency name contains sub.
func (s signals) hasDep(sub string) bool { return anyContains(s.deps, sub) }

// hasFramework reports whether any framework string contains sub.
func (s signals) hasFramework(sub string) bool { return anyContains(s.frames, sub) }

// hasIncoming reports whether a graph edge points into the given (lowercased)
// package path — i.e. the package is a consumed surface that owns behavior.
func (s signals) hasIncoming(pkg string) bool { return s.inEdges[strings.ToLower(pkg)] }

func anyContains(hay []string, sub string) bool {
	for _, h := range hay {
		if strings.Contains(h, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
