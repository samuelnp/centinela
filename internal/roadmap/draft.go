package roadmap

// IsDraftFeature reports whether the named feature is a draft — an authored,
// not-yet-scored feature living in a schedulable phase. Source of truth is the
// persisted Feature.Draft flag; every draft reader consults it independently.
func IsDraftFeature(r *Roadmap, name string) bool {
	if r == nil {
		return false
	}
	for _, p := range r.Phases {
		for _, f := range p.Features {
			if f.Name == name {
				return f.Draft
			}
		}
	}
	return false
}

// DraftFeatures returns every draft Feature across all phases in declared order.
func DraftFeatures(r *Roadmap) []Feature {
	if r == nil {
		return nil
	}
	var out []Feature
	for _, p := range r.Phases {
		for _, f := range p.Features {
			if f.Draft {
				out = append(out, f)
			}
		}
	}
	return out
}
