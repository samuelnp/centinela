package importgraph

import (
	"strings"
	"testing"
)

func TestExecRunner_Success(t *testing.T) {
	out, err := execRunner("echo", "hello")
	if err != nil || !strings.Contains(string(out), "hello") {
		t.Fatalf("got %q %v", out, err)
	}
}

func TestExecRunner_NonZeroFoldsStderr(t *testing.T) {
	_, err := execRunner("sh", "-c", "echo boom 1>&2; exit 1")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("nonzero exit must fold the first stderr line: %v", err)
	}
}

func TestExecRunner_MissingBinary(t *testing.T) {
	if _, err := execRunner("definitely-not-a-real-binary-xyz"); err == nil {
		t.Fatal("a missing binary must error")
	}
}

func TestFirstStderrLine(t *testing.T) {
	if got := firstStderrLine("\n\n  oops \nmore"); got != "oops" {
		t.Fatalf("first non-empty trimmed line, got %q", got)
	}
	if got := firstStderrLine("   \n  "); got != "" {
		t.Fatalf("all-blank -> empty, got %q", got)
	}
}

func TestOnPath(t *testing.T) {
	if !onPath("sh") {
		t.Fatal("sh should resolve on PATH")
	}
	if onPath("definitely-not-a-real-binary-xyz") {
		t.Fatal("a bogus name must not resolve")
	}
}
