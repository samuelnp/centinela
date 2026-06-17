package githooks

import "strings"

// splice replaces the marker-delimited region in existing with block, or
// appends block when no markers are present. It is idempotent: re-splicing the
// identical block over an already-spliced file returns changed=false. Non-marked
// lines are preserved verbatim.
func splice(existing, block string) (result string, changed bool) {
	start := strings.Index(existing, BeginMarker)
	end := strings.Index(existing, EndMarker)
	if start == -1 || end == -1 || end < start {
		return appendBlock(existing, block), true
	}
	end += len(EndMarker)
	// Consume a trailing newline so the region swap stays clean.
	if end < len(existing) && existing[end] == '\n' {
		end++
	}
	prefix := existing[:start]
	suffix := existing[end:]
	next := prefix + block + suffix
	return next, next != existing
}

// appendBlock joins block onto existing, ensuring exactly one blank-line gap
// when existing already has content.
func appendBlock(existing, block string) string {
	if strings.TrimSpace(existing) == "" {
		return block
	}
	base := strings.TrimRight(existing, "\n")
	return base + "\n\n" + block
}

// removeBlock deletes the marker-delimited region (and orphaned blank lines)
// from existing. It reports whether anything was removed and the cleaned text.
func removeBlock(existing string) (result string, changed bool) {
	start := strings.Index(existing, BeginMarker)
	end := strings.Index(existing, EndMarker)
	if start == -1 || end == -1 || end < start {
		return existing, false
	}
	end += len(EndMarker)
	if end < len(existing) && existing[end] == '\n' {
		end++
	}
	cleaned := strings.TrimRight(existing[:start], "\n") + "\n" + existing[end:]
	cleaned = strings.TrimLeft(cleaned, "\n")
	return cleaned, true
}
