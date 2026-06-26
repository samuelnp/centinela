package delivery

import "strings"

// ExtractSection returns the trimmed body of the Markdown `## <heading>`
// section in src: every line after the heading up to (but excluding) the next
// `## ` heading or EOF. The heading match is case-insensitive and ignores
// surrounding whitespace. A missing heading yields "" so the caller can omit
// the dependent section instead of fabricating content.
func ExtractSection(src, heading string) string {
	want := strings.ToLower(strings.TrimSpace(heading))
	if want == "" || src == "" {
		return ""
	}
	lines := strings.Split(src, "\n")
	start := -1
	for i, ln := range lines {
		if h, ok := headingText(ln); ok && strings.ToLower(h) == want {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return ""
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if _, ok := headingText(lines[i]); ok {
			end = i
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines[start:end], "\n"))
}

// headingText reports whether ln is a top-level `## ` heading and returns its
// trimmed title. Deeper headings (`### `) and non-headings return false.
func headingText(ln string) (string, bool) {
	t := strings.TrimSpace(ln)
	if !strings.HasPrefix(t, "## ") {
		return "", false
	}
	if strings.HasPrefix(t, "### ") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(t, "## ")), true
}

// FirstParagraph returns the first blank-line-delimited paragraph of body,
// trimmed. It is used to keep summary sections concise.
func FirstParagraph(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if idx := strings.Index(body, "\n\n"); idx != -1 {
		return strings.TrimSpace(body[:idx])
	}
	return body
}
