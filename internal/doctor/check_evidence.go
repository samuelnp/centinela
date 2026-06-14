package doctor

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/samuelnp/centinela/internal/evidence"
)

// evidenceCheck sweeps orphaned `*.json.tmp` files left under .workflow/ by
// crashed atomic writes, across ALL features. The repair reuses evidence.Repair
// per-feature-prefix; it is safe and idempotent (re-running finds nothing).
type evidenceCheck struct{}

func (evidenceCheck) Name() string { return "evidence" }

func (evidenceCheck) Run(Context) Diagnosis {
	d := Diagnosis{Name: "evidence"}
	tmps := orphanedTmps()
	if len(tmps) == 0 {
		d.Status = OK
		d.Message = "no orphaned evidence temp files"
		return d
	}
	d.Status = Error
	d.Message = "orphaned evidence *.json.tmp files from a crashed write"
	d.Details = append(d.Details, tmps...)
	d.Repair = &Repair{Safe: true, Idempotent: true, Apply: repairEvidence}
	return d
}

// orphanedTmps returns the sorted paths of all *.json.tmp files in .workflow/.
func orphanedTmps() []string {
	matches, _ := filepath.Glob(filepath.Join(".workflow", "*.json.tmp"))
	sort.Strings(matches)
	return matches
}

// repairEvidence removes every orphaned temp via evidence.Repair, grouped by
// the feature prefix derived from each file name (<feature>-<role>.json.tmp).
func repairEvidence() error {
	seen := map[string]bool{}
	for _, p := range orphanedTmps() {
		feature := featurePrefix(filepath.Base(p))
		if feature == "" || seen[feature] {
			continue
		}
		seen[feature] = true
		if _, err := evidence.Repair(feature); err != nil {
			return err
		}
	}
	return nil
}

// featurePrefix extracts the feature slug from a "<feature>-<role>.json.tmp"
// base name. Roles themselves can contain hyphens (e.g. "senior-engineer"), so
// the prefix is found by stripping a known role suffix rather than splitting on
// the last hyphen. Falls back to the bare name when no known role matches.
func featurePrefix(base string) string {
	name := strings.TrimSuffix(base, ".json.tmp")
	for _, r := range evidence.AllRoles() {
		suffix := "-" + string(r)
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix)
		}
	}
	return name
}
