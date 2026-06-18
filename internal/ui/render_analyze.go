package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/analyze"
)

// RenderInventorySummary renders a concise human summary of an analyze
// Inventory: the primary language, the dominant build/test signal, and the
// locale / package / graph-edge counts. cmd/ calls this so analyze itself never
// imports ui.
func RenderInventorySummary(inv analyze.Inventory) string {
	primary := inv.PrimaryLanguage
	if primary == "" {
		primary = "(none)"
	}
	build, test := buildTestSignal(inv.Manifests)
	lines := []string{
		StyleBold.Render("Codebase inventory") +
			StyleMuted.Render(fmt.Sprintf("  (schema v%d)", inv.SchemaVersion)),
		"  primary language: " + primary,
		"  build: " + orNone(build) + "   test: " + orNone(test),
		fmt.Sprintf("  locales: %d   packages: %d   graph edges: %d (%s)",
			len(inv.Locales), len(inv.Packages), len(inv.Graph.Edges), inv.Graph.Kind),
	}
	if inv.Graph.Note != "" {
		lines = append(lines, StyleMuted.Render("  graph note: "+inv.Graph.Note))
	}
	return strings.Join(lines, "\n")
}

// buildTestSignal returns the first non-empty build and test signal across the
// detected manifests (deterministic: manifests are pre-sorted by path).
func buildTestSignal(manifests []analyze.Manifest) (build, test string) {
	for _, m := range manifests {
		if build == "" {
			build = m.Build
		}
		if test == "" {
			test = m.Test
		}
	}
	return build, test
}

// orNone renders an empty signal as a muted "(none)".
func orNone(s string) string {
	if s == "" {
		return StyleMuted.Render("(none)")
	}
	return s
}
