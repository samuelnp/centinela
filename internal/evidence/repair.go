package evidence

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/workflow"
)

// Repair removes orphaned `<feature>-<role>.json.tmp` files left behind by a
// crashed atomic write. It is idempotent and safe to re-run; the caller may
// invoke it before every `set`/`append` flow without risk. Glob errors
// surface only on malformed patterns — the literal pattern below cannot
// trip that path, so the error return is wrapped defensively only.
func Repair(feature string) ([]string, error) {
	prefix := feature + "-"
	matches, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, prefix+"*.json"+tempSuffix))
	removed := []string{}
	for _, path := range matches {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return removed, fmt.Errorf("evidence repair remove %s: %w", path, err)
		}
		removed = append(removed, path)
	}
	return removed, nil
}

// SchemaSkeleton returns the JSON skeleton for the given role rendered for
// embedding in prompts. The CLI version is stamped so the file shows what
// binary produced the skeleton.
func SchemaSkeleton(role Role, cliVersion string) ([]byte, error) {
	skel := Skeleton("<feature-slug>", role, cliVersion)
	return skel.MarshalJSON()
}
