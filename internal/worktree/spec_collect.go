package worktree

import (
	"os"
	"path/filepath"
	"strings"
)

// mainOwner labels scenarios read from the primary checkout.
const mainOwner = "main"

// worktreeSpecsDir is the specs directory carried by one feature worktree.
func worktreeSpecsDir(repo, feature string) string {
	return filepath.Join(repo, Dir, feature, "specs")
}

// mainSpecIndex is the baseline every worktree branched from. It is a
// reference point, never a comparison peer: a difference between main and a
// branch is supersession, which the merge exists to apply.
func mainSpecIndex(repo string) map[string]scenarioRecord {
	return indexByKey(readSpecsFrom(filepath.Join(repo, "specs"), mainOwner))
}

// ownedSpecs pairs a worktree's feature slug with the scenarios it carries.
type ownedSpecs struct {
	owner string
	recs  []scenarioRecord
}

// otherWorktreeSpecs returns the scenarios of every in-flight worktree except
// the one being merged.
func otherWorktreeSpecs(repo, mergingFeature string) []ownedSpecs {
	roots, _ := filepath.Glob(filepath.Join(repo, Dir, "*"))
	var out []ownedSpecs
	for _, root := range roots {
		owner := filepath.Base(root)
		if owner == mergingFeature {
			continue
		}
		recs := readSpecsFrom(filepath.Join(root, "specs"), owner)
		if len(recs) > 0 {
			out = append(out, ownedSpecs{owner: owner, recs: recs})
		}
	}
	return out
}

// readSpecsFrom parses every .feature file in dir into scenarioRecords.
// Returns nil when the directory does not exist.
func readSpecsFrom(dir, owner string) []scenarioRecord {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var recs []scenarioRecord
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".feature") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		recs = append(recs, parseScenarios(string(data), owner, e.Name())...)
	}
	return recs
}
