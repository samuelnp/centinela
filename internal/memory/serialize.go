package memory

import (
	"fmt"
	"strings"
	"time"
)

// marshal renders an Entry as a markdown file with a frontmatter header.
func marshal(e Entry) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", e.ID)
	fmt.Fprintf(&b, "feature: %s\n", e.Feature)
	fmt.Fprintf(&b, "step: %s\n", e.Step)
	fmt.Fprintf(&b, "type: %s\n", e.Type)
	fmt.Fprintf(&b, "title: %s\n", e.Title)
	fmt.Fprintf(&b, "tags: %s\n", strings.Join(e.Tags, ", "))
	fmt.Fprintf(&b, "sourceArtifact: %s\n", e.SourceArtifact)
	fmt.Fprintf(&b, "createdAt: %s\n", e.CreatedAt.UTC().Format(time.RFC3339))
	b.WriteString("---\n\n")
	b.WriteString(e.Body)
	b.WriteString("\n")
	return []byte(b.String())
}

// unmarshal parses a ledger entry file back into an Entry.
func unmarshal(data []byte) (Entry, bool) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return Entry{}, false
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Entry{}, false
	}
	e := parseFrontmatter(rest[:end])
	body := strings.TrimSpace(rest[end+len("\n---"):])
	e.Body = strings.TrimPrefix(body, "\n")
	e.Body = strings.TrimSpace(e.Body)
	if e.ID == "" {
		return Entry{}, false
	}
	return e, true
}

func parseFrontmatter(block string) Entry {
	var e Entry
	for _, line := range strings.Split(block, "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		assignField(&e, strings.TrimSpace(k), strings.TrimSpace(v))
	}
	return e
}

func assignField(e *Entry, key, val string) {
	switch key {
	case "id":
		e.ID = val
	case "feature":
		e.Feature = val
	case "step":
		e.Step = val
	case "type":
		e.Type = val
	case "title":
		e.Title = val
	case "tags":
		e.Tags = splitTags(val)
	case "sourceArtifact":
		e.SourceArtifact = val
	case "createdAt":
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			e.CreatedAt = t
		}
	}
}

func splitTags(val string) []string {
	out := []string{}
	for _, t := range strings.Split(val, ",") {
		if tt := strings.TrimSpace(t); tt != "" {
			out = append(out, tt)
		}
	}
	return out
}
