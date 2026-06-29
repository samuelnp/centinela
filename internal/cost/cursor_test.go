package cost

import (
	"os"
	"testing"
)

// chdir into a temp dir so the fixed cursor path is sandboxed per test.
func inTempRepo(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestCursorRoundTrip(t *testing.T) {
	inTempRepo(t)
	if c := LoadCursor(); c.Path != "" || c.Offset != 0 {
		t.Fatalf("absent cursor should be zero, got %+v", c)
	}
	SaveCursor("/tmp/t.jsonl", 1234)
	c := LoadCursor()
	if c.Path != "/tmp/t.jsonl" || c.Offset != 1234 {
		t.Fatalf("round-trip mismatch: %+v", c)
	}
}

func TestOffsetForMatchesPathOnly(t *testing.T) {
	c := Cursor{Path: "/a.jsonl", Offset: 99}
	if got := c.OffsetFor("/a.jsonl"); got != 99 {
		t.Fatalf("same path should return saved offset, got %d", got)
	}
	if got := c.OffsetFor("/b.jsonl"); got != 0 {
		t.Fatalf("new session (different path) should reset to 0, got %d", got)
	}
}

func TestLoadCursorMalformedIsZero(t *testing.T) {
	inTempRepo(t)
	_ = os.MkdirAll(cursorDir, 0o755)
	_ = os.WriteFile(cursorFile, []byte("not json"), 0o644)
	if c := LoadCursor(); c.Path != "" || c.Offset != 0 {
		t.Fatalf("malformed cursor should read as zero, got %+v", c)
	}
}
