package workflow

import (
	"fmt"
	"strings"
)

// Archetypes are named start-time presets that select a subset and ordering of
// the canonical step names — never new step identifiers. Every step in every
// archetype is a canonical name already in the gating matrix, so the existing
// step-gating, required-role, per-step validator and ship-gate mechanisms work
// unchanged. Archetype (sequence) is orthogonal to the enforcement profile
// (strictness).
const (
	ArchetypeCanonical = "canonical"
	ArchetypeHotfix    = "hotfix"
	ArchetypeRefactor  = "refactor"
	ArchetypeSpike     = "spike"
)

// NormalizeArchetype trims and lowercases an archetype name. An empty value
// resolves to canonical (the default track); a known value passes through.
// An UNKNOWN value passes through UNCHANGED so ValidateArchetype can reject it
// — coercing a typo to canonical would silently run the wrong track.
func NormalizeArchetype(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	if v == "" {
		return ArchetypeCanonical
	}
	return v
}

// DisplayArchetype returns the archetype name to surface in read-only views and
// an optional annotation. The pinned per-feature value is used, defaulting to
// canonical when unset; spike is annotated to make its missing ship gate visible.
func DisplayArchetype(wf *Workflow) (name, note string) {
	name = ArchetypeCanonical
	if wf != nil && wf.Archetype != "" {
		name = NormalizeArchetype(wf.Archetype)
	}
	if name == ArchetypeSpike {
		note = "spike — no ship gate"
	}
	return name, note
}

// ValidateArchetype accepts an empty value (resolves to canonical) or one of the
// four known archetypes; any other value is a configuration error naming the
// offending value and the "archetype" field.
func ValidateArchetype(name string) error {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", ArchetypeCanonical, ArchetypeHotfix, ArchetypeRefactor, ArchetypeSpike:
		return nil
	default:
		return fmt.Errorf("archetype %q is unsupported (use canonical, hotfix, refactor, or spike)", name)
	}
}
