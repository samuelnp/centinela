package reconstruct

import (
	"strings"
)

// featureSkeleton assembles a role-aware Gherkin .feature body for a target as a
// pure, byte-stable string. It emits a Feature: line, a derived narrative, and
// exactly one Scenario: whose Given/When/Then are explicit "# TODO: confirm"
// markers — never a fabricated concrete assertion. The shape satisfies the real
// spec_traceability parser (a Feature: line + ≥1 indented Scenario: line). It
// returns the body and the count of TODO markers it contains.
func featureSkeleton(t Target) (body string, todos int) {
	tpl := templateFor(t.Role)
	var b strings.Builder
	b.WriteString("Feature: " + t.Slug + " — reconstructed " + string(roleOrModule(t.Role)) + " behavior\n")
	b.WriteString("  As a maintainer of " + t.Pkg + "\n")
	b.WriteString("  I want to " + narrativeFor(t.Role) + "\n")
	b.WriteString("  So that the reconstructed skeleton can be confirmed against real behavior\n\n")
	b.WriteString("  # Reconstructed from the analyze Inventory: " + t.Reason + ".\n")
	b.WriteString("  # Every step below is a placeholder — replace each marker with confirmed behavior.\n")
	b.WriteString("  Scenario: " + tpl.name + "\n")
	b.WriteString("    Given " + tpl.given + " " + todoMarker + "\n")
	b.WriteString("    When " + tpl.when + " " + todoMarker + "\n")
	b.WriteString("    Then " + tpl.then + " " + todoMarker + "\n")
	return b.String(), strings.Count(b.String(), todoMarker)
}

// roleOrModule normalizes an empty role to RoleModule for display.
func roleOrModule(r Role) Role {
	if r == "" {
		return RoleModule
	}
	return r
}
