package delivery

import "strings"

// ComposeChangelog produces one ChangelogEntry from e. The seed line is the
// first non-blank, non-FILL line of the changelog stub; if the stub yields
// nothing usable the line is derived from the brief's title/Problem. The
// category is chosen from the seed's conventional-commit prefix:
// feat → Added, fix → Fixed, everything else → Changed.
func ComposeChangelog(e Evidence) ChangelogEntry {
	seed := stubSeed(e.ChangelogStub)
	if seed == "" {
		seed = deriveFromBrief(e)
	}
	return ChangelogEntry{Category: categoryFor(seed), Line: bulletize(seed)}
}

// stubSeed returns the first stub line that is non-blank, not a heading, and
// not still a FILL slot. Returns "" when none qualifies.
func stubSeed(stub string) string {
	for _, raw := range strings.Split(stub, "\n") {
		ln := strings.TrimSpace(raw)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		if strings.Contains(strings.ToUpper(ln), "FILL") {
			continue
		}
		return strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(ln, "-")), " ")
	}
	return ""
}

// deriveFromBrief builds a fallback line "<feature>: <summary>" from the
// brief's Problem section, or just the feature slug when absent.
func deriveFromBrief(e Evidence) string {
	summary := FirstParagraph(ExtractSection(e.Brief, "Problem"))
	summary = strings.TrimSpace(strings.SplitN(summary, "\n", 2)[0])
	if summary == "" {
		return e.Feature
	}
	return e.Feature + ": " + summary
}

// categoryFor maps a conventional-commit-style prefix to a Keep-a-Changelog
// subsection, defaulting to Changed.
func categoryFor(seed string) string {
	low := strings.ToLower(strings.TrimSpace(seed))
	switch {
	case strings.HasPrefix(low, "feat") || strings.HasPrefix(low, "feature"):
		return "Added"
	case strings.HasPrefix(low, "fix") || strings.HasPrefix(low, "bug"):
		return "Fixed"
	default:
		return "Changed"
	}
}

// bulletize normalizes seed into a single trimmed bullet text (no leading "-").
func bulletize(seed string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(seed), "-"))
}
