package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

// hookContextInput is the slice of the UserPromptSubmit payload this hook reads.
type hookContextInput struct {
	SessionID string `json:"session_id"`
}

// readHookSessionID drains stdin (avoiding SIGPIPE on large prompts) and
// best-effort parses the host session id. An unreadable stream, a non-JSON
// body and a payload without session_id all yield "" — the fail-open signal
// that forces a render.
func readHookSessionID(r io.Reader) string {
	data, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	var in hookContextInput
	if err := json.Unmarshal(data, &in); err != nil {
		return ""
	}
	return in.SessionID
}

// emitRoadmapSummary prints the one-line roadmap summary only when it is new
// for this session or its projection changed. It owns no decision of its own —
// every branch is a domain call — and every error path degrades to rendering
// or to silence, never to an error surfaced to the host session.
func emitRoadmapSummary(sessionID string) {
	r, err := roadmap.Load()
	if err != nil {
		return // unchanged behavior: no roadmap line and no digest write
	}
	path := roadmap.SummaryStatePath()
	digest := roadmap.SummaryDigest(r)
	if !roadmap.ShouldRenderSummary(roadmap.LoadSummaryState(path), sessionID, digest) {
		return
	}
	fmt.Println(ui.RenderRoadmapSummary(r))
	//nolint:errcheck // an unwritable .workflow/ must never break the host session
	_ = roadmap.SaveSummaryState(path, roadmap.SummaryState{SessionID: sessionID, Digest: digest})
}
