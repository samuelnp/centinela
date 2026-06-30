// Package telemetry is the governance event log: a non-blocking, append-only
// JSONL recorder (one Event per line) plus a lenient reader. It is a config-only
// leaf — it imports internal/config and stdlib only, never internal/workflow,
// internal/verify, or internal/gates (it owns its own CheckRef copy). Producers
// in domain packages stay pure; emission happens from cmd/ call-sites.
package telemetry

// Schema is the versioned event identifier on every line (greppable, matches
// the centinela.verdict/v1 convention). Bump only on a breaking schema change.
const Schema = "centinela.telemetry/v1"

const (
	TypeBlock            = "block"
	TypeGateFailure      = "gate-failure"
	TypeVerifyRejection  = "verify-rejection"
	TypeCompleteRejected = "complete-rejected"
	TypeStepAdvanced     = "step-advanced"
	TypeStepRevised      = "step-revised"
	TypeCostSample       = "cost-sample"
)

// Event is one governance event, serialized as a single JSON line in
// .workflow/telemetry/events.jsonl. Flat + omitempty so the 5 downstream
// readers can consume it stably. The only nested type is CheckRef (owned here,
// NOT imported from internal/verify, to keep this package a config-only leaf).
type Event struct {
	Schema       string     `json:"schema"`
	Type         string     `json:"type"`
	Timestamp    string     `json:"timestamp"` // RFC3339 UTC
	Feature      string     `json:"feature,omitempty"`
	Step         string     `json:"step,omitempty"`
	From         string     `json:"from,omitempty"`         // step-revised: the step rewound away from (Step holds the target)
	Model        string     `json:"model,omitempty"`        // driver model id, stamped at emit (back-compat: old lines → "")
	Reason       string     `json:"reason,omitempty"`       // block: need-init|out-of-step ; complete-rejected: gates|verify
	FileType     string     `json:"fileType,omitempty"`     // block
	TargetPath   string     `json:"targetPath,omitempty"`   // block
	Gate         string     `json:"gate,omitempty"`         // gate-failure
	Message      string     `json:"message,omitempty"`      // gate-failure
	Checks       []CheckRef `json:"checks,omitempty"`       // verify-rejection
	InputTokens  int        `json:"inputTokens,omitempty"`  // cost-sample (back-compat: old lines → 0)
	OutputTokens int        `json:"outputTokens,omitempty"` // cost-sample
}

// CheckRef is a failing claim check (telemetry's own copy of verify's shape).
type CheckRef struct {
	Claim  string `json:"claim"`
	Role   string `json:"role"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}
