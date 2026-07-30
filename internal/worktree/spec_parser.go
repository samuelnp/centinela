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
// Given, and Then lines per scenario block. It is good enough to compare two
// copies of the same scenario; it is NOT a full Gherkin parser.
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
