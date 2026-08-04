package gatereport

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// FenceOpen is the tagged fence that opens the verification block. Tagging it
// makes parsing unambiguous — an untagged ```json fence elsewhere in the
// report is ignored, and no prose is ever scanned.
const FenceOpen = "```json centinela:verification"

// ErrNoBlock reports an absent verification fence. Callers translate it into
// the operator-facing "no commands-run record" remedy.
var ErrNoBlock = errors.New("report has no centinela:verification block")

// ParseVerification returns the record carried by the FIRST tagged fence.
//
// The raw `commands` array is schema-checked BEFORE it is decoded into the
// typed slice, because decoding is lossy in the direction that matters: a
// missing exitCode becomes 0 (= passed) and an entry that is not an object at
// all vanishes. The same rule runs at stamp time (Stamped), so read and write
// agree on one shape.
func ParseVerification(report string) (Verification, error) {
	body, ok := blockBody(report)
	if !ok {
		return Verification{}, ErrNoBlock
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &fields); err != nil {
		return Verification{}, fmt.Errorf("malformed centinela:verification block: %w", err)
	}
	if err := ValidateCommandsSchema(fields["commands"]); err != nil {
		return Verification{}, fmt.Errorf("malformed centinela:verification block: %w", err)
	}
	var v Verification
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return Verification{}, fmt.Errorf("malformed centinela:verification block: %w", err)
	}
	return v, nil
}

// blockBody returns the JSON text between the tagged fence and its closer.
func blockBody(report string) (string, bool) {
	start, end, ok := blockSpan(report)
	if !ok {
		return "", false
	}
	return report[start:end], true
}

// blockSpan locates the JSON body of the first tagged fence as a byte range
// into report, exclusive of both fence lines. An unterminated fence is not a
// block. Offsets (not just the text) so the stamp rewrite can splice in place.
func blockSpan(report string) (start, end int, ok bool) {
	pos, bodyStart := 0, -1
	for _, ln := range strings.SplitAfter(report, "\n") {
		trimmed := strings.TrimSpace(ln)
		next := pos + len(ln)
		switch {
		case bodyStart < 0 && trimmed == FenceOpen:
			bodyStart = next
		case bodyStart >= 0 && trimmed == "```":
			return bodyStart, pos, true
		}
		pos = next
	}
	return 0, 0, false
}
