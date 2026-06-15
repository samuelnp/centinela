package calibration

import (
	"sort"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
)

// Calibrate turns a parsed event slice into a deterministic per-model Report. It
// is pure (no I/O) and stdlib-only beyond the telemetry + config leaves. An empty
// slice yields an empty-state Report (ModelCount=0, Models=[], empty span).
func Calibrate(events []telemetry.Event, cfg *config.Config) Report {
	start, end := span(events)
	stats := frictionByModel(events)

	models := make([]ModelRecord, 0, len(stats))
	for id, s := range stats {
		verdict, rec, recProfile, class, current := classify(id, s, cfg)
		models = append(models, ModelRecord{
			Model:              id,
			Class:              class,
			CurrentProfile:     current,
			Friction:           s,
			Recommendation:     rec,
			RecommendedProfile: recProfile,
			Verdict:            verdict,
		})
	}
	sortModels(models)

	return Report{
		ModelCount: len(models),
		SpanStart:  start,
		SpanEnd:    end,
		Models:     models,
	}
}

// sortModels orders by model id ascending with "unattributed" forced last, for
// byte-identical output across runs (never ranges a map in output order).
func sortModels(m []ModelRecord) {
	sort.Slice(m, func(i, j int) bool {
		ui, uj := m[i].Model == unattributed, m[j].Model == unattributed
		if ui != uj {
			return uj // a real id sorts before "unattributed"
		}
		return m[i].Model < m[j].Model
	})
}

// span returns the earliest and latest event timestamps (RFC3339 lexicographic),
// or ("","") for an empty slice.
func span(events []telemetry.Event) (start, end string) {
	for _, e := range events {
		if e.Timestamp == "" {
			continue
		}
		if start == "" || e.Timestamp < start {
			start = e.Timestamp
		}
		if end == "" || e.Timestamp > end {
			end = e.Timestamp
		}
	}
	return start, end
}
