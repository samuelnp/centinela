// Package verdict aggregates gates, verify, evidence, and workflow provenance
// into a deterministic, machine-readable JSON packet. It is Lipgloss-free and
// injects its timestamp (no time.Now inside) so output is byte-stable for fixed
// inputs. It is an aggregator package and stays unmapped in the import graph.
package verdict

// Packet is the top-level centinela.verdict/v1 document.
type Packet struct {
	Schema   string      `json:"schema"`
	Run      RunInfo     `json:"run"`
	Summary  Summary     `json:"summary"`
	Gates    []GateLine  `json:"gates"`
	Verify   []CheckLine `json:"verify"`
	Evidence []EvidLine  `json:"evidence"`
}

// RunInfo snapshots the workflow and config provenance for the run.
type RunInfo struct {
	Feature     string `json:"feature"`
	Step        string `json:"step"`
	Profile     string `json:"profile"`
	Archetype   string `json:"archetype"`
	DriverModel string `json:"driverModel,omitempty"`
	Headless    bool   `json:"headless"`
	GeneratedAt string `json:"generatedAt"`
}

// Summary is the computed pass/fail verdict plus tallies.
type Summary struct {
	Verdict  string `json:"verdict"`  // "pass" | "fail"
	ExitCode int    `json:"exitCode"` // 0 | 1
	Gates    Counts `json:"gates"`
	Verify   Counts `json:"verify"`
}

// Counts tallies outcomes by category. A struct (not a map) keeps marshaling
// deterministic.
type Counts struct {
	Pass int `json:"pass"`
	Fail int `json:"fail"`
	Warn int `json:"warn"`
	Skip int `json:"skip"`
}

// GateLine is one gate result; status is lowercased (pass/fail/warn/skip).
type GateLine struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Message string   `json:"message,omitempty"`
	Details []string `json:"details,omitempty"`
}

// CheckLine is one verify claim check; status keeps verify's native UPPERCASE
// vocabulary (PASS/FAIL/SKIP/WARN/CONFIG-ERROR/TIMEOUT).
type CheckLine struct {
	Role   string `json:"role"`
	Claim  string `json:"claim"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// EvidLine is one on-disk role evidence entry.
type EvidLine struct {
	Role        string `json:"role"`
	Step        string `json:"step"`
	Status      string `json:"status"`
	HandoffTo   string `json:"handoffTo,omitempty"`
	GeneratedAt string `json:"generatedAt,omitempty"`
	Path        string `json:"path"`
}
