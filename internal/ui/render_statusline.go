package ui

import "strings"

type StatusLineView struct {
	Primary   []string
	Secondary []string
}

func RenderStatusLine(v StatusLineView) string {
	line1 := strings.TrimSpace(strings.Join(v.Primary, " "))
	line2 := strings.TrimSpace(strings.Join(v.Secondary, " "))
	if line1 == "" && line2 == "" {
		return ""
	}
	if line2 == "" {
		return line1
	}
	if line1 == "" {
		return line2
	}
	return line1 + "\n" + line2
}
