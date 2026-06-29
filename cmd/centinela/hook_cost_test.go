package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// feedStdin replaces os.Stdin with a pipe carrying payload for the duration of fn.
func feedStdin(t *testing.T, payload string, fn func()) {
	t.Helper()
	r, w, _ := os.Pipe()
	old := os.Stdin
	os.Stdin = r
	_, _ = w.WriteString(payload)
	_ = w.Close()
	fn()
	os.Stdin = old
}

func TestHookCostCapturesAndAttributes(t *testing.T) {
	dir := seedCostRepo(t)
	// Replace the seeded log so we observe only the hook's own write.
	_ = os.Remove(filepath.Join(dir, ".workflow", "telemetry", "events.jsonl"))
	tp := filepath.Join(dir, "t.jsonl")
	_ = os.WriteFile(tp, []byte(
		`{"message":{"usage":{"input_tokens":700,"output_tokens":300}}}`+"\n"), 0o644)

	feedStdin(t, `{"cwd":"`+dir+`","transcript_path":"`+tp+`"}`, func() {
		if err := runHookCost(nil, nil); err != nil {
			t.Fatal(err)
		}
	})

	events, _ := telemetry.ReadDefault()
	var samples int
	for _, e := range events {
		if e.Type == telemetry.TypeCostSample {
			samples++
			if e.Feature != "demo" || e.Step != "code" || e.InputTokens != 700 || e.OutputTokens != 300 {
				t.Fatalf("bad attribution: %+v", e)
			}
		}
	}
	if samples != 1 {
		t.Fatalf("want 1 cost sample, got %d", samples)
	}
}

func TestHookCostNoTranscriptIsNoOp(t *testing.T) {
	seedCostRepo(t)
	feedStdin(t, `{"cwd":"x"}`, func() {
		if err := runHookCost(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	// seeded log still has exactly its one pre-existing sample (no new write)
	events, _ := telemetry.ReadDefault()
	if len(events) != 1 {
		t.Fatalf("no transcript_path → no new sample, got %d events", len(events))
	}
}

func TestHookCostDisabledIsNoOp(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "centinela.toml"), []byte(""), 0o644)
	t.Chdir(dir)
	tp := filepath.Join(dir, "t.jsonl")
	_ = os.WriteFile(tp, []byte(`{"message":{"usage":{"input_tokens":9}}}`+"\n"), 0o644)
	feedStdin(t, `{"cwd":"`+dir+`","transcript_path":"`+tp+`"}`, func() {
		_ = runHookCost(nil, nil)
	})
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "telemetry", "events.jsonl")); !strings.Contains(errString(err), "no such") {
		t.Fatal("cost disabled should write no telemetry")
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
