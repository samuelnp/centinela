package migration

import "strings"

func extractCustomSections(current, template string) []string {
	t := sectionMap(template)
	c := sectionMap(current)
	var out []string
	for title, body := range c {
		if _, ok := t[title]; ok || title == "Preserved Custom Sections" {
			continue
		}
		if strings.HasPrefix(title, "Preserved Keep Blocks") {
			continue
		}
		out = append(out, body)
	}
	return out
}

func sectionMap(markdown string) map[string]string {
	lines := strings.Split(markdown, "\n")
	out := map[string]string{}
	start := -1
	title := ""
	for i, line := range lines {
		if !strings.HasPrefix(line, "## ") {
			continue
		}
		if start >= 0 {
			out[title] = strings.Join(lines[start:i], "\n")
		}
		title = strings.TrimSpace(strings.TrimPrefix(line, "## "))
		start = i
	}
	if start >= 0 {
		out[title] = strings.Join(lines[start:], "\n")
	}
	return out
}

func appendCustomSections(base string, sections []string) string {
	if len(sections) == 0 {
		return base
	}
	extra := "\n\n## Preserved Custom Sections\n\n" + strings.Join(sections, "\n\n")
	return strings.TrimRight(base, "\n") + extra + "\n"
}
