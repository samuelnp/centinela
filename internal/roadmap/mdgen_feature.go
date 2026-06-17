package roadmap

import "strings"

// renderFeature renders a normal feature bullet. The base is "- **<Name>**";
// a non-empty Description is appended after " — "; a non-empty DependsOn adds
// a single " (depends on a, b)" clause in declared slice order (attached
// directly to the bullet when there is no description). A non-empty Fixes adds
// an indented "  *Fixes: …*" line beneath the bullet.
func renderFeature(f Feature) []string {
	line := "- **" + f.Name + "**"
	if f.Description != "" {
		line += " — " + f.Description
	}
	if len(f.DependsOn) > 0 {
		line += " (depends on " + strings.Join(f.DependsOn, ", ") + ")"
	}
	out := []string{line}
	if f.Fixes != "" {
		out = append(out, "  *Fixes: "+f.Fixes+"*")
	}
	return out
}

// renderBacklogFeature renders a deferred-finding line:
// "- **<Name>** — <Summary> *(deferred <DeferredAt> · <Feature>/<Role>)*".
// Empty parenthetical parts are omitted; the parenthetical is dropped entirely
// when no provenance fields are present, never emitting an empty "()".
func renderBacklogFeature(f Feature) []string {
	line := "- **" + f.Name + "**"
	if f.Summary != "" {
		line += " — " + f.Summary
	}
	if paren := backlogParenthetical(f); paren != "" {
		line += " *(" + paren + ")*"
	}
	return []string{line}
}

// backlogParenthetical builds the " deferred <at> · <feature>/<role>" body,
// omitting whichever parts are empty so no "· /" or empty group is emitted.
func backlogParenthetical(f Feature) string {
	var parts []string
	if f.DeferredAt != "" {
		parts = append(parts, "deferred "+f.DeferredAt)
	}
	if prov := backlogProvenance(f.Source); prov != "" {
		parts = append(parts, prov)
	}
	return strings.Join(parts, " · ")
}

// backlogProvenance renders "<feature>/<role>" from a Source, omitting an empty
// half so it never emits a bare "/".
func backlogProvenance(s *Source) string {
	if s == nil || (s.Feature == "" && s.Role == "") {
		return ""
	}
	if s.Feature == "" {
		return s.Role
	}
	if s.Role == "" {
		return s.Feature
	}
	return s.Feature + "/" + s.Role
}
