package delivery

import "strings"

var canonicalOrder = []string{"Added", "Changed", "Fixed"}

// InsertEntry inserts entry's bullet under `### <Category>` inside the
// `## [Unreleased]` block of changelogMD and returns the new text plus whether
// it changed. It is pure and idempotent: when the normalized bullet already
// exists anywhere in the [Unreleased] block it returns the text unchanged and
// false. A missing `### <Category>` subsection is created in the canonical
// order Added → Changed → Fixed. Released sections are never touched.
func InsertEntry(changelogMD string, entry ChangelogEntry) (string, bool) {
	lines := strings.Split(changelogMD, "\n")
	start, end := unreleasedBounds(lines)
	if start == -1 {
		return changelogMD, false
	}
	bullet := "- " + bulletize(entry.Line)
	if blockHasBullet(lines[start:end], bullet) {
		return changelogMD, false
	}
	out := insertBullet(lines, start, end, entry.Category, bullet)
	return strings.Join(out, "\n"), true
}

// unreleasedBounds returns the line index just after `## [Unreleased]` and the
// exclusive end of its block (next `## ` heading, `---` rule, or EOF).
func unreleasedBounds(lines []string) (int, int) {
	start := -1
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "## [Unreleased]") {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return -1, -1
	}
	for i := start; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "## ") || t == "---" {
			return start, i
		}
	}
	return start, len(lines)
}

// blockHasBullet reports whether any line in block equals bullet after trimming
// trailing whitespace.
func blockHasBullet(block []string, bullet string) bool {
	want := strings.TrimRight(bullet, " \t")
	for _, ln := range block {
		if strings.TrimRight(ln, " \t") == want {
			return true
		}
	}
	return false
}

// insertBullet places bullet at the end of the `### <category>` subsection
// within [start,end), creating the subsection in canonical order if absent.
func insertBullet(lines []string, start, end int, category, bullet string) []string {
	if at := subsectionEnd(lines, start, end, category); at != -1 {
		return splice(lines, at, []string{bullet})
	}
	at := newSubsectionAt(lines, start, end, category)
	return splice(lines, at, []string{"### " + category, bullet, ""})
}
