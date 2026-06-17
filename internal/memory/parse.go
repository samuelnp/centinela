package memory

import (
	"fmt"
	"strings"
	"time"
)

// parseLesson harvests edge-case lessons (tests step) into a single entry.
func parseLesson(feature, source, text string, at time.Time) ([]Entry, error) {
	body := collectBullets(text)
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("no edge-case lessons found")
	}
	tags := []string{"edge-cases", "lesson"}
	return []Entry{newEntry(feature, "tests", TypeLesson, body, source, tags, at)}, nil
}

// parseVerdict harvests the gatekeeper verdict (validate step) into one entry.
func parseVerdict(feature, source, text string, at time.Time) ([]Entry, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("empty gatekeeper report")
	}
	tags := []string{"gatekeeper", "verdict"}
	return []Entry{newEntry(feature, "validate", TypeVerdict, strings.TrimSpace(text), source, tags, at)}, nil
}

// parseDecisions harvests one decision entry per bullet under a "## Decisions"
// section (plan step). A missing section yields no entries and no error (SC-04).
func parseDecisions(feature, source, text string, at time.Time) ([]Entry, error) {
	bullets := decisionBullets(text)
	out := []Entry{}
	for _, b := range bullets {
		tags := []string{"decision"}
		out = append(out, newEntry(feature, "plan", TypeDecision, b, source, tags, at))
	}
	return out, nil
}

// decisionBullets returns each bullet under the first "## Decisions" heading.
func decisionBullets(text string) []string {
	lines := strings.Split(text, "\n")
	out := []string{}
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			inSection = strings.EqualFold(strings.TrimSpace(trimmed[3:]), "Decisions")
			continue
		}
		if !inSection {
			continue
		}
		if bullet := bulletText(trimmed); bullet != "" {
			out = append(out, bullet)
		}
	}
	return out
}

func bulletText(line string) string {
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:])
	}
	return ""
}

func collectBullets(text string) string {
	out := []string{}
	for _, line := range strings.Split(text, "\n") {
		if b := bulletText(strings.TrimSpace(line)); b != "" {
			out = append(out, "- "+b)
		}
	}
	return strings.Join(out, "\n")
}
