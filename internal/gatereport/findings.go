package gatereport

import "strings"

// FirstFinding returns the first bullet under the report's "Findings" heading,
// stripped of list and emphasis markers, so a blocking verdict can echo the
// reason instead of a bare severity token. Returns a placeholder when the
// report carries no findings — a CRITICAL verdict still blocks either way.
func FirstFinding(report string) string {
	inFindings := false
	for _, ln := range strings.Split(report, "\n") {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "#") {
			inFindings = strings.HasPrefix(strings.TrimLeft(trimmed, "# "), "Findings")
			continue
		}
		if !inFindings {
			continue
		}
		if bullet := stripBullet(trimmed); bullet != "" {
			return bullet
		}
	}
	return "see the report's Findings section"
}

func stripBullet(line string) string {
	if !strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "* ") {
		return ""
	}
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(line[2:]), "*"))
}
