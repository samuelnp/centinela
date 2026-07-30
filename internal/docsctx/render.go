package docsctx

import (
	"fmt"
	"strings"
)

// Render assembles the docs context as plain markdown for stdout —
// pipe-friendly, no UI panels. Each section embeds its file verbatim under
// a `> source:` line; an absent changelog draft prints an actionable hint
// instead of a body.
func Render(ctx Context) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Docs Context: %s\n", ctx.Feature)
	writeSection(&b, "Feature brief", ctx.Brief)
	writeSection(&b, "Plan", ctx.Plan)
	writeSection(&b, "Spec scenarios", ctx.Spec)
	writeChangelog(&b, ctx)
	return b.String()
}

func writeSection(b *strings.Builder, title string, s Section) {
	fmt.Fprintf(b, "\n## %s\n\n> source: %s\n\n%s\n", title, s.Source, strings.TrimRight(s.Body, "\n"))
}

func writeChangelog(b *strings.Builder, ctx Context) {
	if ctx.Changelog.Present {
		writeSection(b, "Changelog draft", ctx.Changelog)
		return
	}
	fmt.Fprintf(b, "\n## Changelog draft\n\n> source: %s\n\nno changelog draft yet — run: centinela artifact new %s changelog\n", ctx.Changelog.Source, ctx.Feature)
}
