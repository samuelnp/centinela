package main

import (
	"os"
	"testing"
)

// Stdin parsing for the hook context payload: every failure mode must yield an
// empty session id, which ShouldRenderSummary turns into a render (AC17).

// errReader always fails, standing in for a stdin stream the host tore down.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, os.ErrClosed }

// AC17: an unreadable stdin yields "" — the fail-open signal that renders.
func TestReadHookSessionIDUnreadableStdin(t *testing.T) {
	if got := readHookSessionID(errReader{}); got != "" {
		t.Fatalf("unreadable stdin must yield an empty session id, got %q", got)
	}
}

func TestReadHookSessionIDFailsOpen(t *testing.T) {
	cases := map[string]string{
		`{"session_id":"s-9"}`:           "s-9",
		`{"session_id":"s-9","x":1}`:     "s-9",
		"":                               "", // E21: empty stdin
		"not json at all":                "", // E21: non-JSON
		`{"other":"field"}`:              "", // E21: no session_id
		`{"session_id":"","cwd":"/tmp"}`: "",
	}
	for payload, want := range cases {
		var got string
		withStdin(t, payload, func() { got = readHookSessionID(os.Stdin) })
		if got != want {
			t.Errorf("readHookSessionID(%q) = %q, want %q", payload, got, want)
		}
	}
}
