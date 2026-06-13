package roadmap

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/workflow"
)

// ValidateDependencies checks that all dependsOn references name known features,
// that every feature archetype is supported, and that no dependency cycle
// exists. Returns nil when the roadmap has no deps.
func ValidateDependencies(r *Roadmap) error {
	if r == nil {
		return nil
	}
	names := roadmapFeatureSet(r)
	deps := map[string][]string{}
	for _, phase := range r.Phases {
		for _, f := range phase.Features {
			if err := workflow.ValidateArchetype(f.Archetype); err != nil {
				return fmt.Errorf("feature %s: %w", f.Name, err)
			}
			for _, dep := range f.DependsOn {
				if !names[dep] {
					return fmt.Errorf(
						"feature %s depends on unknown feature %s",
						f.Name, dep,
					)
				}
			}
			deps[f.Name] = f.DependsOn
		}
	}
	if hasCycle(deps) {
		return fmt.Errorf("roadmap dependency cycle detected")
	}
	return nil
}
