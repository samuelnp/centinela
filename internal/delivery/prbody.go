package delivery

import "strings"

// ComposePRBody assembles the PR body Markdown from e, emitting these sections
// in order and OMITTING (not faking) any whose source datum is absent:
//
//	Summary → What changed / Why → Acceptance → Gate status → provenance footer.
//
// The provenance footer is always present (constant text only). The function is
// pure: it performs no I/O.
func ComposePRBody(e Evidence) string {
	sections := []string{
		summarySection(e),
		whatWhySection(e),
		acceptanceSection(e),
		gateStatusSection(e),
	}
	out := make([]string, 0, len(sections)+1)
	for _, s := range sections {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	out = append(out, provenanceFooter(e))
	return strings.Join(out, "\n\n") + "\n"
}

// ComposePRTitle returns a concise PR title for e: the changelog summary line
// (a conventional-commit-shaped one-liner), or the feature slug as a fallback.
// `gh pr create` requires a title when non-interactive, so this must never be
// empty.
func ComposePRTitle(e Evidence) string {
	if t := strings.TrimSpace(ComposeChangelog(e).Line); t != "" {
		return t
	}
	return e.Feature
}
