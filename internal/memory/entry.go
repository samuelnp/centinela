// Package memory implements the governed project-memory ledger: capturing
// typed facts from step artifacts and recalling the relevant slice at plan time.
package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Entry type discriminators.
const (
	TypeLesson   = "lesson"
	TypeVerdict  = "verdict"
	TypeDecision = "decision"
)

// Entry is one git-tracked fact in the ledger.
type Entry struct {
	ID             string
	Feature        string
	Step           string
	Type           string
	Title          string
	Tags           []string
	SourceArtifact string
	CreatedAt      time.Time
	Body           string
}

// computeID derives a stable content hash from the dedupe-relevant fields.
// Source path is excluded so the same fact across worktrees stays identical.
func computeID(feature, typ, body string) string {
	sum := sha256.Sum256([]byte(feature + "\x00" + typ + "\x00" + strings.TrimSpace(body)))
	return hex.EncodeToString(sum[:8])
}

// newEntry builds an Entry with a derived ID and a one-line title.
func newEntry(feature, step, typ, body, source string, tags []string, at time.Time) Entry {
	return Entry{
		ID:             computeID(feature, typ, body),
		Feature:        feature,
		Step:           step,
		Type:           typ,
		Title:          firstLine(body),
		Tags:           tags,
		SourceArtifact: source,
		CreatedAt:      at,
		Body:           strings.TrimSpace(body),
	}
}

func firstLine(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}

func (e Entry) sizeBytes() int { return len(e.Title) + len(e.Body) + len(fmt.Sprint(e.Tags)) }
