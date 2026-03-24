package roadmap

import "strings"

func HasBootstrapPhase(r *Roadmap) bool {
	return len(BootstrapFeatures(r)) > 0
}

func IsBootstrapFeature(r *Roadmap, feature string) bool {
	for _, name := range BootstrapFeatures(r) {
		if name == feature {
			return true
		}
	}
	return false
}

func BootstrapComplete(r *Roadmap) bool {
	features := BootstrapFeatures(r)
	if len(features) == 0 {
		return false
	}
	for _, f := range features {
		if FeatureStatus(f) != "done" {
			return false
		}
	}
	return true
}

func BootstrapFeatures(r *Roadmap) []string {
	if r == nil {
		return nil
	}
	var names []string
	for _, phase := range r.Phases {
		if !isBootstrapPhaseName(phase.Name) {
			continue
		}
		for _, feature := range phase.Features {
			names = append(names, feature.Name)
		}
	}
	return names
}

func isBootstrapPhaseName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return strings.HasPrefix(n, "phase 0") && strings.Contains(n, "bootstrap")
}
