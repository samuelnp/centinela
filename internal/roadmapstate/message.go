package roadmapstate

import "strings"

// SubjectLimit caps the subject half of a roadmap commit message so a long
// slug or phase name cannot produce an unreadable git log entry.
const SubjectLimit = 60

// Message builds the conventional commit message for a roadmap mutation:
// "chore(roadmap): <verb> <subject>". The subject is trimmed, collapsed to a
// single line (a phase note may contain newlines) and truncated to
// SubjectLimit runes with an ellipsis. An empty subject yields just the verb.
func Message(verb, subject string) string {
	msg := "chore(roadmap): " + strings.TrimSpace(verb)
	s := truncate(singleLine(subject), SubjectLimit)
	if s == "" {
		return msg
	}
	return msg + " " + s
}

// singleLine collapses every run of whitespace (newlines included) to one space.
func singleLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// truncate shortens s to at most limit RUNES, ending in "…" when it cut. Runes,
// not bytes, so a multi-byte subject is never sliced mid-character.
func truncate(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit-1]) + "…"
}
