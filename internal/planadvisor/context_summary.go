package planadvisor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func contextLines(b bundle) []string {
	lines := []string{}
	if len(b.Dependencies) > 0 {
		lines = append(lines, "- dependencies first: "+strings.Join(b.Dependencies, ", "))
	}
	if len(b.Siblings) > 0 {
		lines = append(lines, "- same-phase siblings: "+strings.Join(b.Siblings, ", "))
	}
	if len(b.Lessons) > 0 {
		lines = append(lines, "- related edge-case lessons: "+strings.Join(b.Lessons, "; "))
	}
	if len(b.QualityNotes) > 0 {
		lines = append(lines, "- roadmap quality notes: "+strings.Join(b.QualityNotes, "; "))
	}
	if len(b.Memory) > 0 {
		lines = append(lines, "- 🛡️👁️ MEMORY (recalled facts): "+strings.Join(b.Memory, "; "))
	}
	if len(b.Failures) > 0 {
		lines = append(lines, "- recurring gate failures: "+failureSummary(b.Failures))
	}
	return lines
}

func relatedLessons(names []string) []string {
	out := []string{}
	for _, name := range names {
		if lesson := firstLesson(readText(fmt.Sprintf(".workflow/%s-edge-cases.md", name))); lesson != "" {
			out = append(out, name+": "+lesson)
		}
		if len(out) == 2 {
			return out
		}
	}
	return out
}

func relatedQualityNotes(feature string, names []string) []string {
	data, err := os.ReadFile(roadmap.RoadmapQualityFile)
	if err != nil {
		return nil
	}
	var q roadmap.QualityReport
	if json.Unmarshal(data, &q) != nil {
		return nil
	}
	allowed := map[string]bool{feature: true}
	for _, name := range names {
		allowed[name] = true
	}
	out := []string{}
	for _, f := range q.Features {
		summary := strings.TrimSpace(f.Summary)
		if allowed[f.Name] && summary != "" && strings.ToLower(summary) != "ok" {
			out = append(out, f.Name+": "+summary)
		}
		if len(out) == 2 {
			return out
		}
	}
	return out
}

func firstLesson(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if trimmed != "" && !strings.HasPrefix(trimmed, "Edge Cases") {
			return trimmed
		}
	}
	return ""
}

func take(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return items[:n]
}
