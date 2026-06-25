package delivery

import "strings"

// subsectionEnd returns the index where a bullet should be appended inside an
// existing `### <category>` subsection, or -1 when the subsection is absent.
func subsectionEnd(lines []string, start, end int, category string) int {
	head := "### " + category
	in := false
	last := -1
	for i := start; i < end; i++ {
		t := strings.TrimSpace(lines[i])
		switch {
		case t == head:
			in, last = true, i+1
		case in && strings.HasPrefix(t, "### "):
			return last
		case in && strings.HasPrefix(t, "- "):
			last = i + 1
		}
	}
	if in {
		return last
	}
	return -1
}

// newSubsectionAt finds the insertion index for a brand-new `### <category>`
// subsection, honoring the canonical Added → Changed → Fixed order.
func newSubsectionAt(lines []string, start, end int, category string) int {
	rank := indexOf(canonicalOrder, category)
	for i := start; i < end; i++ {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "### ") {
			name := strings.TrimSpace(strings.TrimPrefix(t, "### "))
			if indexOf(canonicalOrder, name) > rank {
				return i
			}
		}
	}
	return end
}

func splice(lines []string, at int, ins []string) []string {
	out := make([]string, 0, len(lines)+len(ins))
	out = append(out, lines[:at]...)
	out = append(out, ins...)
	return append(out, lines[at:]...)
}

func indexOf(xs []string, s string) int {
	for i, x := range xs {
		if x == s {
			return i
		}
	}
	return len(xs)
}
