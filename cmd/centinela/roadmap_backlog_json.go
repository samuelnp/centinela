package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// backlogPayload is the machine shape Magallanes' Plan page consumes. The
// counts always describe the WHOLE Backlog; --stale narrows findings only, so
// a client can show "13 of 106" without a second call.
type backlogPayload struct {
	ThresholdDays int              `json:"threshold_days"`
	Total         int              `json:"total"`
	Stale         int              `json:"stale"`
	Findings      []backlogFinding `json:"findings"`
}

// backlogFinding is one deferred finding. AgeDays is null — not zero — when
// deferredAt is missing or unparseable, so a client can never mistake an
// unreadable clock for a brand-new finding.
type backlogFinding struct {
	Slug       string          `json:"slug"`
	Summary    string          `json:"summary"`
	Source     *roadmap.Source `json:"source"`
	DeferredAt string          `json:"deferredAt"`
	AgeDays    *int            `json:"ageDays"`
	Stale      bool            `json:"stale"`
}

// emitBacklogJSON writes the payload to stdout as indented JSON.
func emitBacklogJSON(rows []roadmap.Aged, stats roadmap.BacklogStats) error {
	payload := backlogPayload{
		ThresholdDays: stats.ThresholdDays,
		Total:         stats.Total,
		Stale:         stats.Stale,
		Findings:      make([]backlogFinding, 0, len(rows)),
	}
	for _, a := range rows {
		payload.Findings = append(payload.Findings, backlogFinding{
			Slug: a.Name, Summary: a.Summary, Source: a.Source,
			DeferredAt: a.DeferredAt, AgeDays: ageOrNil(a), Stale: a.Stale,
		})
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, string(data))
	return nil
}

// ageOrNil returns a pointer to the age, or nil when the age is unknown.
func ageOrNil(a roadmap.Aged) *int {
	if !a.KnownAge {
		return nil
	}
	days := a.AgeDays
	return &days
}
