package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/gates"
)

// RenderGateResult renders a single gate result line.
func RenderGateResult(r gates.Result) string {
	switch r.Status {
	case gates.Pass:
		return StyleGreen.Render(IconDone+" "+r.Name) + "  " + StyleMuted.Render(r.Message)
	case gates.Fail:
		lines := []string{StyleRed.Render("✗ " + r.Name) + "  " + r.Message}
		for _, d := range r.Details {
			lines = append(lines, StyleMuted.Render("  · "+d))
		}
		return strings.Join(lines, "\n")
	case gates.Warn:
		return StyleYellow.Render("⚠ "+r.Name) + "  " + StyleMuted.Render(r.Message)
	case gates.Skip:
		return StyleMuted.Render("— "+r.Name+"  "+r.Message)
	}
	return ""
}

// RenderCmdResult renders the result of a user-defined validate command.
func RenderCmdResult(cmd string, passed bool, output string) string {
	var icon, label string
	if passed {
		icon = StyleGreen.Render(IconDone)
		label = StyleGreen.Render(cmd)
	} else {
		icon = StyleRed.Render("✗")
		label = StyleRed.Render(cmd)
	}
	line := fmt.Sprintf("%s  %s", icon, label)
	if !passed && strings.TrimSpace(output) != "" {
		line += "\n" + indentBlock(output)
	}
	return line
}

func indentBlock(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		b.WriteString(StyleMuted.Render("  │ ") + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
