package memory

import (
	"fmt"
	"os"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

type parser func(feature, source, text string, at time.Time) ([]Entry, error)

type sourceSpec struct {
	path  string
	parse parser
}

// sourceFor maps a just-completed step to its capture source. Steps with no
// capture source (code, docs) return false.
func sourceFor(feature, step string) (sourceSpec, bool) {
	switch step {
	case "tests":
		return sourceSpec{fmt.Sprintf(".workflow/%s-edge-cases.md", feature), parseLesson}, true
	case "validate":
		return sourceSpec{fmt.Sprintf(".workflow/%s-gatekeeper.md", feature), parseVerdict}, true
	case "plan":
		return sourceSpec{fmt.Sprintf("docs/features/%s.md", feature), parseDecisions}, true
	}
	return sourceSpec{}, false
}

// Capture harvests the artifact for the just-completed step into the ledger.
// It never returns an error: failures are warnings so they never block the
// workflow advance (SC-04/06/07). Disabled config is a no-op (SC-12).
func Capture(feature, step string, cfg *config.Config) {
	if cfg == nil || !cfg.Memory.IsEnabled() {
		return
	}
	spec, ok := sourceFor(feature, step)
	if !ok {
		return
	}
	data, err := os.ReadFile(spec.path)
	if err != nil {
		warn("skipping %s — %v", spec.path, err)
		return
	}
	entries, err := spec.parse(feature, spec.path, string(data), time.Now())
	if err != nil {
		warn("skipping %s — %v", spec.path, err)
		return
	}
	if len(entries) == 0 {
		return
	}
	persist(entries)
}

func persist(entries []Entry) {
	for _, e := range entries {
		if _, err := writeIfAbsent(e); err != nil {
			warn("cannot write entry %s — %v", e.ID, err)
		}
	}
	if err := regenerateIndex(); err != nil {
		warn("cannot regenerate index — %v", err)
	}
}

func warn(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[memory] warning: "+format+"\n", args...)
}
