package workflow

import "strings"

func hasAcceptanceExecutionCommand(commands []string) bool {
	for _, cmd := range commands {
		c := strings.ToLower(strings.TrimSpace(cmd))
		if c == "" {
			continue
		}
		if strings.Contains(c, "tests/acceptance") {
			return true
		}
		if strings.Contains(c, "go test") && strings.Contains(c, "./...") {
			return true
		}
		if strings.Contains(c, "cucumber") || strings.Contains(c, "godog") || strings.Contains(c, "behave") {
			return true
		}
		if strings.Contains(c, "acceptance") && strings.Contains(c, "test") {
			return true
		}
	}
	return false
}
