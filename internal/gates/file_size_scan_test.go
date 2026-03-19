package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanHelpers(t *testing.T) {
	if !shouldSkipDir(".git") || !shouldSkipDir(".hidden") || shouldSkipDir("src") {
		t.Fatal("shouldSkipDir mismatch")
	}
	if !isSourceFile("x.go") || !isSourceFile("x.ts") || isSourceFile("x.txt") {
		t.Fatal("isSourceFile mismatch")
	}
	if itoa(0) != "0" || itoa(42) != "42" {
		t.Fatal("itoa mismatch")
	}
}

func TestCountLinesAndFormat(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "a.go")
	os.WriteFile(p, []byte("a\nb\nc\n"), 0644) //nolint:errcheck
	if n := countLines(p); n != 3 {
		t.Fatalf("countLines=%d", n)
	}
	v := formatViolation(p, 3)
	if v == "" || !strings.Contains(v, "(3 lines)") {
		t.Fatalf("bad violation format: %q", v)
	}
}
