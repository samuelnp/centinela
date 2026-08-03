// Acceptance: specs/dynamic-model-routing.feature
//
// Shared binary + fixture helpers for the dynamic-model-routing suite. The
// binary is built ONCE (sync.Once) from ./cmd/centinela into a temp dir — never
// the installed binary, which predates `route`. Every fixture is a local temp
// directory; no test in this suite touches the network.
package acceptance_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var dmrBinOnce sync.Once
var dmrBin string

func dmrBuildBin(t *testing.T) string {
	t.Helper()
	dmrBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-dmr-bin")
		if err != nil {
			t.Fatal(err)
		}
		dmrBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", dmrBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return dmrBin
}

const dmrDynamicTOML = "[orchestration]\nrouting_mode = \"dynamic\"\n"
const dmrStaticTOML = "[orchestration]\nrouting_mode = \"static\"\n"

// dmrStartedAt is fixed in the past so any artifact a test writes counts as
// belonging to THIS run (see the stale-evidence guard).
const dmrStartedAt = "2026-01-01T00:00:00Z"

// dmrRepo creates a temp project with .workflow/ and the given centinela.toml.
// No docs/features/<f>.md is written, so features are NOT user-facing and the
// ux-ui-specialist / documentation-specialist roles are never scheduled.
func dmrRepo(t *testing.T, toml string) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: existing\n")
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if toml != "" {
		mustWrite(t, filepath.Join(dir, "centinela.toml"), toml)
	}
	return dir
}

// dmrWorkflowJSON is a strict, contract-pinned canonical workflow. extra is
// appended raw (e.g. `,"modelRoutes":{…}`) so a test can hand-write state the
// way an agent with write access to .workflow/*.json would.
func dmrWorkflowJSON(feature, step, extra string) string {
	return fmt.Sprintf(`{"feature":%q,"startedAt":%q,"currentStep":%q,`+
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
		`"orchestrationMode":"strict-subagents-v1","planContract":"planner-v1",`+
		`"validateContract":"adversarial-v1"%s}`, feature, dmrStartedAt, step, extra)
}

// dmrLegacyWorkflowJSON predates BOTH pinned contracts and modelRoutes: its plan
// step schedules big-thinker + feature-specialist and its validate step
// schedules validation-specialist.
func dmrLegacyWorkflowJSON(feature, step string, extra string) string {
	return fmt.Sprintf(`{"feature":%q,"startedAt":%q,"currentStep":%q,`+
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
		`"orchestrationMode":"strict-subagents-v1"%s}`, feature, dmrStartedAt, step, extra)
}

func dmrWrite(t *testing.T, dir, feature, body string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"), body)
}

// dmrState returns the raw workflow JSON on disk.
func dmrState(t *testing.T, dir, feature string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".workflow", feature+".json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	return string(data)
}

// dmrEvidence writes an evidence stub for a role (delegation has begun).
func dmrEvidence(t *testing.T, dir, feature, role string) string {
	t.Helper()
	path := filepath.Join(dir, ".workflow", feature+"-"+role+".md")
	mustWrite(t, path, "# stub\n")
	return path
}

// dmrOK runs the binary and fails the test when it does not exit 0.
func dmrOK(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	out, code := runCent(t, bin, dir, args...)
	if code != 0 {
		t.Fatalf("run %v exit %d: %s", args, code, out)
	}
	return out
}

// dmrRefused runs the binary expecting a non-zero exit and returns the output.
func dmrRefused(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	out, code := runCent(t, bin, dir, args...)
	if code == 0 {
		t.Fatalf("run %v must be refused, got exit 0: %s", args, out)
	}
	return out
}
