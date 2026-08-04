package orchestration

import "strings"

// StubEntry reports whether a document entry is still scaffold rather than
// something an author wrote. It is UnreplacedSlot plus the case that rule alone
// cannot see: an entry built ENTIRELY out of citation slots.
//
// "- <FILL: ...>: <FILL: ...>" carries no real description, so every slot in it
// reads as a citation and UnreplacedSlot passes it — yet blanking the scaffold's
// two descriptions to ellipses is not writing a changelog, it is erasing the
// evidence that one was never written. Prose that legitimately QUOTES the marker
// says something either side of it; a stub does not.
func StubEntry(entry string) bool {
	if UnreplacedSlot(entry) {
		return true
	}
	return hasSlot(entry) && !hasSubstanceOutsideSlots(entry)
}

func hasSlot(s string) bool { return strings.Contains(s, fillPrefix) }

// hasSubstanceOutsideSlots reports whether anything survives removing every
// rendered slot but list bullets, separators, and whitespace.
func hasSubstanceOutsideSlots(s string) bool {
	var b strings.Builder
	for rest := s; ; {
		i := strings.Index(rest, fillPrefix)
		if i < 0 {
			b.WriteString(rest)
			break
		}
		b.WriteString(rest[:i])
		rest = rest[i+len(fillPrefix):]
		end := strings.Index(rest, ">")
		if end < 0 {
			b.WriteString(rest)
			break
		}
		rest = rest[end+1:]
	}
	return strings.Trim(strings.TrimSpace(b.String()), "-*:;,.…`\"' \t") != ""
}
