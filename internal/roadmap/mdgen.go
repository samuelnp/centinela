package roadmap

import "strings"

// RenderMarkdown deterministically renders a Roadmap to ROADMAP.md bytes.
// It iterates only ordered slices (Phases, Features, DependsOn) — never a Go
// map — so output is byte-stable across platforms. The result uses LF line
// endings and ends with exactly one trailing newline.
func RenderMarkdown(r *Roadmap) ([]byte, error) {
	var lines []string
	lines = append(lines, "# Roadmap", "")
	if r.Intro != "" {
		lines = append(lines, renderBlockquote(r.Intro)...)
		lines = append(lines, "")
	}
	for _, phase := range r.Phases {
		lines = append(lines, renderPhase(phase)...)
		lines = append(lines, "")
	}
	// Collapse the trailing blank line into the single EOF newline.
	out := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	return []byte(out + "\n"), nil
}

// renderBlockquote renders a prose string as a Markdown blockquote: each line is
// prefixed with "> ", and a blank line inside the prose becomes a bare ">" so
// the blockquote stays unbroken (matching the authored ROADMAP.md style).
func renderBlockquote(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			out = append(out, ">")
			continue
		}
		out = append(out, "> "+line)
	}
	return out
}
