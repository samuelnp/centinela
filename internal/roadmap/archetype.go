package roadmap

// FeatureArchetype returns the archetype pinned on a feature in the roadmap, or
// the empty string when the feature is absent or has no archetype set. An empty
// result lets the start-time resolver fall through to its next precedence tier.
func FeatureArchetype(r *Roadmap, feature string) string {
	if r == nil {
		return ""
	}
	for _, phase := range r.Phases {
		for _, f := range phase.Features {
			if f.Name == feature {
				return f.Archetype
			}
		}
	}
	return ""
}
