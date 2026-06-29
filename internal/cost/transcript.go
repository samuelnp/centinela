// Package cost is a read-only cost-governance analytics aggregator over the
// telemetry log plus the host-harness transcript. It imports internal/telemetry
// and internal/config (leaves) and stdlib only; it is imported solely by cmd/
// (its Report type by internal/ui for rendering). It NEVER blocks: every path
// degrades to zero on missing/malformed input.
package cost

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

// transcriptLine is the tolerant shape of one host-harness JSONL transcript
// line. Token usage may sit under message.usage (Claude Code) or a top-level
// usage; both are summed. Unknown shapes contribute zero.
type transcriptLine struct {
	Message struct {
		Usage tokenUsage `json:"usage"`
	} `json:"message"`
	Usage tokenUsage `json:"usage"`
}

type tokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// SumFrom reads <path> starting at byte <offset>, sums input/output token usage
// across well-formed lines, and returns the new end offset to persist as the
// cursor. A missing file or any read error yields zeros and the unchanged
// offset with no error — capture must never fail the host command. Malformed
// lines are skipped.
func SumFrom(path string, offset int64) (in, out int, newOffset int64, err error) {
	f, oerr := os.Open(path)
	if oerr != nil {
		if errors.Is(oerr, fs.ErrNotExist) {
			return 0, 0, offset, nil
		}
		return 0, 0, offset, nil
	}
	defer f.Close() //nolint:errcheck

	size, _ := f.Seek(0, 2)
	if offset > size { // transcript rotated/truncated → recount from start
		offset = 0
	}
	if _, serr := f.Seek(offset, 0); serr != nil {
		return 0, 0, offset, nil
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var l transcriptLine
		if json.Unmarshal(sc.Bytes(), &l) != nil {
			continue
		}
		in += l.Message.Usage.InputTokens + l.Usage.InputTokens
		out += l.Message.Usage.OutputTokens + l.Usage.OutputTokens
	}
	return in, out, size, nil
}
