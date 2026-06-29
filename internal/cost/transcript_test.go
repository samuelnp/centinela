package cost

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTranscript(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "t.jsonl")
	data := ""
	for _, l := range lines {
		data += l + "\n"
	}
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSumFromSumsAndSkipsGarbage(t *testing.T) {
	p := writeTranscript(t,
		`{"type":"assistant","message":{"usage":{"input_tokens":1200,"output_tokens":800}}}`,
		`not json {`,
		`{"usage":{"input_tokens":300,"output_tokens":200}}`, // top-level usage
	)
	in, out, off, err := SumFrom(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	if in != 1500 || out != 1000 {
		t.Fatalf("got in=%d out=%d, want 1500/1000", in, out)
	}
	if off == 0 {
		t.Fatal("expected non-zero end offset")
	}
}

func TestSumFromMissingFileIsNoOp(t *testing.T) {
	in, out, off, err := SumFrom(filepath.Join(t.TempDir(), "absent.jsonl"), 42)
	if in != 0 || out != 0 || off != 42 || err != nil {
		t.Fatalf("missing file should be a no-op, got %d %d %d %v", in, out, off, err)
	}
}

func TestSumFromOffsetReadsOnlyDelta(t *testing.T) {
	p := writeTranscript(t, `{"message":{"usage":{"input_tokens":100,"output_tokens":0}}}`)
	_, _, off, _ := SumFrom(p, 0)
	in, out, _, _ := SumFrom(p, off) // nothing new past the cursor
	if in != 0 || out != 0 {
		t.Fatalf("delta read should be empty, got %d/%d", in, out)
	}
}

func TestSumFromTruncationResetsToStart(t *testing.T) {
	p := writeTranscript(t, `{"message":{"usage":{"input_tokens":50,"output_tokens":0}}}`)
	// offset beyond EOF (file rotated/truncated) → recount from 0
	in, _, _, _ := SumFrom(p, 1<<20)
	if in != 50 {
		t.Fatalf("truncation should recount from start, got %d", in)
	}
}
