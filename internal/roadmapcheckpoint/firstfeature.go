package roadmapcheckpoint

import "github.com/samuelnp/centinela/internal/roadmap"

// FirstIncompleteBootstrap returns the first Phase 0 bootstrap feature whose
// derived status is not "done", walking BootstrapFeatures in declared order.
// The second return value is false when no bootstrap features exist or all
// of them are already done (the caller should then suppress the directive).
func FirstIncompleteBootstrap(r *roadmap.Roadmap) (string, bool) {
	if r == nil {
		return "", false
	}
	for _, name := range roadmap.BootstrapFeatures(r) {
		if first, ok := roadmap.FirstNotDone(name); ok {
			return first, true
		}
	}
	return "", false
}
