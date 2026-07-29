package gatereport

import "strings"

// Command is one command the verifier actually executed. argv is recorded
// verbatim so the record can later be cross-checked against ground truth.
type Command struct {
	Argv       []string `json:"argv"`
	ExitCode   int      `json:"exitCode"`
	DurationMS int      `json:"durationMs,omitempty"`
}

// Line renders argv as a shell-ish string for matching and error messages.
func (c Command) Line() string {
	return strings.Join(c.Argv, " ")
}

// Verification is the machine-readable grounding record carried inside the
// report's ```json centinela:verification fence (D2). Revision + TreeDigest
// bind the verdict to the exact tree state that was verified (D3).
type Verification struct {
	Revision   string    `json:"revision"`
	TreeDigest string    `json:"treeDigest"`
	Commands   []Command `json:"commands"`
}
