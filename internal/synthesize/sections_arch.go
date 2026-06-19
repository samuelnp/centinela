package synthesize

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/analyze"
)

// renderArchitecture renders the Architecture Choice block from the inferred
// archetype's profile.
func renderArchitecture(inf Inference, p profile) string {
	return fmt.Sprintf(`## Architecture Choice

**Archetype:** %s

**Pattern:** %s

**Why:** <!-- TODO: confirm — inferred from the signals in the rationale above (confidence %s). -->

**Reference:** %s

**G2 rule (layer boundaries):** %s

**G7 rule (outer layer):** %s`,
		inf.Best, p.pattern, inf.Confidence, p.reference, p.g2, p.g7)
}

// renderLayerMapping derives each abstract layer's concrete path by matching the
// archetype's layer keywords against the inventory packages. An unmatched slot
// emits a TODO so the draft is honest about gaps.
func renderLayerMapping(inv analyze.Inventory, p profile) string {
	var rows strings.Builder
	for _, slot := range p.layers {
		rows.WriteString(fmt.Sprintf("| %s | %s |\n", slot.name, matchPaths(inv.Packages, slot.keyword)))
	}
	return "## Layer Mapping\n\n| Abstract Layer | Concrete Path |\n|---------------|---------------|\n" +
		strings.TrimRight(rows.String(), "\n")
}

// renderGatekeeper lists the paths the gatekeeper should scan: feature specs
// plus each non-empty inferred layer path.
func renderGatekeeper(inv analyze.Inventory, p profile) string {
	var rows strings.Builder
	rows.WriteString("| Feature specs | `specs/` |\n")
	for _, slot := range p.layers {
		if m := matchPaths(inv.Packages, slot.keyword); !strings.HasPrefix(m, "<!--") {
			rows.WriteString(fmt.Sprintf("| %s | %s |\n", slot.name, m))
		}
	}
	return "## Gatekeeper Paths\n\n| What | Path |\n|------|------|\n" + strings.TrimRight(rows.String(), "\n")
}

// matchPaths returns the packages whose path contains keyword as inline-coded
// paths, or a TODO placeholder when none match.
func matchPaths(pkgs []string, keyword string) string {
	var hits []string
	for _, p := range pkgs {
		if strings.Contains(strings.ToLower(p), keyword) {
			hits = append(hits, "`"+p+"`")
		}
	}
	if len(hits) == 0 {
		return "<!-- TODO: confirm -->"
	}
	return strings.Join(hits, ", ")
}
