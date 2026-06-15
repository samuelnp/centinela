// Package calibration computes read-only per-model governance-friction analysis
// over the telemetry event log and recommends a tighter, looser, or unchanged
// enforcement profile per driver model. It is an aggregator over two leaves: it
// imports only internal/telemetry + internal/config + stdlib, never cmd/ or
// internal/ui. Calibrate is pure and deterministic — models are sorted (id asc,
// "unattributed" last); no map is ranged in output order. Output is advisory: a
// Report only, never an auto-applied config write.
package calibration

// Verdict classifies how well a model's current profile fits its measured
// friction. JSON-serialized via its String() — a stable --json contract.
type Verdict string

const (
	Undergoverned  Verdict = "Undergoverned"
	Overgoverned   Verdict = "Overgoverned"
	WellCalibrated Verdict = "WellCalibrated"
	Unclassified   Verdict = "Unclassified"
)

// Recommendation is the advised action for a model's profile.
type Recommendation string

const (
	Tighten Recommendation = "Tighten"
	Loosen  Recommendation = "Loosen"
	Keep    Recommendation = "Keep"
	None    Recommendation = "None"
)

// Report is the pure, serializable calibration payload. Field names are a stable
// --json contract; do not rename without bumping consumers. Models is sorted
// deterministically (id asc, "unattributed" last).
type Report struct {
	ModelCount int           // distinct model buckets (incl. "unattributed")
	SpanStart  string        // earliest event timestamp (RFC3339), "" if none
	SpanEnd    string        // latest event timestamp, "" if none
	Models     []ModelRecord // per-model record, deterministically ordered
}

// ModelRecord is one model's friction evidence, current/recommended profile, and
// classification. Carries the raw counts so every recommendation is auditable.
type ModelRecord struct {
	Model              string
	Class              string
	CurrentProfile     string
	Friction           FrictionStats
	Recommendation     Recommendation
	RecommendedProfile string
	Verdict            Verdict
}

// FrictionStats are the raw counts + derived rate driving classification.
// Rework = GateFailures + VerifyRejections + complete-rejected. Rate =
// Rework / Advances (friction per successful advance); HasRate=false guards
// Advances==0 (no division, no NaN).
type FrictionStats struct {
	Blocks           int
	GateFailures     int
	VerifyRejections int
	Rework           int
	Advances         int
	Rate             float64
	HasRate          bool
}
