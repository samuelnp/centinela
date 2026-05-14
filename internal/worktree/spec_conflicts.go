package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SpecConflict names two feature spec files and the scenario whose Given
// clauses diverge between them.
type SpecConflict struct {
	FeatureA string
	FeatureB string
	Scenario string
	Given    string
}

// DetectSpecConflicts loads `specs/*.feature` from the main checkout and from
// each active worktree, and reports scenarios where the same Given context
// asserts different observable outcomes. mergingFeature is the feature about
// to be merged — it is included in the comparison set.
func DetectSpecConflicts(repo, mergingFeature string) []SpecConflict {
	scenarios := collectScenarios(repo, mergingFeature)
	return scenariosConflicts(scenarios)
}

// FormatSpecConflicts produces a human-readable list of conflicts.
func FormatSpecConflicts(conflicts []SpecConflict) string {
	lines := make([]string, 0, len(conflicts))
	for _, c := range conflicts {
		lines = append(lines, fmt.Sprintf("%s ↔ %s — scenario %q (Given: %s)",
			c.FeatureA, c.FeatureB, c.Scenario, c.Given))
	}
	return strings.Join(lines, "; ")
}

// collectScenarios reads every .feature it can reach for the comparison set.
func collectScenarios(repo, mergingFeature string) []scenarioRecord {
	var out []scenarioRecord
	out = append(out, readSpecsFrom(filepath.Join(repo, "specs"), "main")...)
	worktreeRoots, _ := filepath.Glob(filepath.Join(repo, Dir, "*"))
	for _, root := range worktreeRoots {
		feat := filepath.Base(root)
		out = append(out, readSpecsFrom(filepath.Join(root, "specs"), feat)...)
	}
	_ = mergingFeature
	return out
}

// readSpecsFrom parses every .feature file in dir into scenarioRecords.
// Returns an empty slice when the directory does not exist.
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
