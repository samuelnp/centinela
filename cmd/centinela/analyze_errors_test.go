package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunAnalyze_UnwritableOutFails(t *testing.T) {
	analyzeRepo(t)
	if err := os.WriteFile("blocker", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runAnalyzeCmd(t, "blocker/inv.json")
	if err == nil || !strings.Contains(err.Error(), "cannot write") {
		t.Fatalf("un-writable out must hard-error: %v", err)
	}
}

func TestRunAnalyze_UnreadableRootFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	// 0o311 dir: chdir succeeds (search) but ReadDir(".") fails (no read), so
	// Analyze(".") hard-errors on the unreadable root.
	dir := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(o) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o311); err != nil {
		t.Skip("cannot restrict dir: " + err.Error())
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	_, err := runAnalyzeCmd(t, "/tmp/should-not-write.json")
	if err == nil || !strings.Contains(err.Error(), "unreadable root") {
		t.Fatalf("unreadable root must hard-error: %v", err)
	}
}
