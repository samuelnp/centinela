package roadmap

// renderPhase renders one phase: its "## " heading (carrying any authored
// status glyph verbatim), an optional note blockquote, then its features. A
// Backlog phase formats each feature as a deferred-finding line; every other
// phase uses the normal feature bullet. A phase with zero features renders just
// its heading (and optional note) without stray blank lines.
func renderPhase(p Phase) []string {
	out := []string{"## " + p.Name, ""}
	if p.Note != "" {
		out = append(out, renderBlockquote(p.Note)...)
		out = append(out, "")
	}
	backlog := isBacklogPhaseName(p.Name)
	for _, f := range p.Features {
		if backlog {
			out = append(out, renderBacklogFeature(f)...)
		} else {
			out = append(out, renderFeature(f)...)
		}
	}
	// Drop the trailing blank emitted by an empty note when the phase has no
	// features, so a noteless empty phase ends cleanly on its heading.
	return trimTrailingBlank(out)
}

// trimTrailingBlank removes a single trailing empty line so phases with an
// optional note but no features don't leave a dangling blank before the
// caller's phase separator.
func trimTrailingBlank(lines []string) []string {
	if n := len(lines); n > 0 && lines[n-1] == "" {
		return lines[:n-1]
	}
	return lines
}
