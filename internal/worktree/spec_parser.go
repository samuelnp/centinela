package worktree

import (
	"bufio"
	"strings"
)

// scenarioRecord captures a single Given/Then pair from a .feature file.
type scenarioRecord struct {
	Owner    string // feature slug or "main"
	File     string
	Scenario string
	Given    string
	Then     string
}

// parseScenarios is a deliberately small Gherkin reader: it pulls Scenario,
// Given, and Then lines per scenario block. It is good enough to flag
// contradictions; it is NOT a full Gherkin parser.
func parseScenarios(text, owner, file string) []scenarioRecord {
	var recs []scenarioRecord
	var cur scenarioRecord
	flush := func() {
		if cur.Scenario != "" {
			cur.Owner = owner
			cur.File = file
			recs = append(recs, cur)
		}
		cur = scenarioRecord{}
	}
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "Scenario:"):
			flush()
			cur.Scenario = strings.TrimSpace(strings.TrimPrefix(line, "Scenario:"))
		case strings.HasPrefix(line, "Given ") && cur.Given == "":
			cur.Given = strings.TrimSpace(strings.TrimPrefix(line, "Given "))
		case strings.HasPrefix(line, "Then ") && cur.Then == "":
			cur.Then = strings.TrimSpace(strings.TrimPrefix(line, "Then "))
		}
	}
	flush()
	return recs
}

// scenariosConflicts groups records by their Given clause; if two records
// share a Given but disagree on Then, that is a conflict.
func scenariosConflicts(recs []scenarioRecord) []SpecConflict {
	byGiven := map[string][]scenarioRecord{}
	for _, r := range recs {
		if r.Given == "" || r.Then == "" {
			continue
		}
		byGiven[r.Given] = append(byGiven[r.Given], r)
	}
	var conflicts []SpecConflict
	for given, group := range byGiven {
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				if group[i].Owner == group[j].Owner {
					continue
				}
				if group[i].Then != group[j].Then {
					conflicts = append(conflicts, SpecConflict{
						FeatureA: group[i].File,
						FeatureB: group[j].File,
						Scenario: group[i].Scenario,
						Given:    given,
					})
				}
			}
		}
	}
	return conflicts
}
