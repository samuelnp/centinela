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
		lines := []string{StyleRed.Render("✗ "+r.Name) + "  " + r.Message}
		for _, d := range r.Details {
			lines = append(lines, StyleMuted.Render("  · "+d))
		}
		return strings.Join(lines, "\n")
	case gates.Warn:
		return StyleYellow.Render("⚠ "+r.Name) + "  " + StyleMuted.Render(r.Message)
	case gates.Skip:
		return StyleMuted.Render("— " + r.Name + "  " + r.Message)
	}
	return ""
}

// RenderCmdVerdict renders one validate command's three-valued verdict. detail
// is the muted one-line reason (empty for an unremarkable pass); the captured
// output is echoed only for a failure, where it is the thing to act on.
func RenderCmdVerdict(cmd string, status gates.Status, detail, output string) string {
	var icon, label string
	switch status {
	case gates.Fail:
		icon, label = StyleRed.Render("✗"), StyleRed.Render(cmd)
	case gates.Warn:
		icon, label = StyleYellow.Render("⚠"), StyleYellow.Render(cmd)
	default:
		icon, label = StyleGreen.Render(IconDone), StyleGreen.Render(cmd)
	}
	line := fmt.Sprintf("%s  %s", icon, label)
	if strings.TrimSpace(detail) != "" {
		line += "  " + StyleMuted.Render(detail)
	}
	if status == gates.Fail && strings.TrimSpace(output) != "" {
		line += "\n" + indentBlock(output)
	}
	return line
}

// RenderCmdResult renders the result of a user-defined validate command. Kept
// as the bool-shaped entry point so existing callers and tests are untouched.
func RenderCmdResult(cmd string, passed bool, output string) string {
	status := gates.Pass
	if !passed {
		status = gates.Fail
	}
	return RenderCmdVerdict(cmd, status, "", output)
}

func indentBlock(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		b.WriteString(StyleMuted.Render("  │ ") + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
