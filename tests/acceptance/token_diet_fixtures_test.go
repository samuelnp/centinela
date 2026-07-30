// Acceptance: specs/token-diet.feature
//
// Shared binary + fixture helpers for the token-diet acceptance suite. The
// binary is built ONCE (sync.Once) for the whole file set, mirroring
// unified_plan_specialist_fixtures_test.go. Reuses the package-level
// mustWrite/runCent/repoRoot/runGit/writeFile helpers already defined
// elsewhere in this package — do not redefine them here.
package acceptance_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var tdBinOnce sync.Once
var tdBin string

// tdBuildBin compiles the centinela binary once for every token-diet test.
func tdBuildBin(t *testing.T) string {
	t.Helper()
	tdBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-td-bin")
		if err != nil {
			t.Fatal(err)
		}
		tdBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", tdBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return tdBin
}

// tdRepo creates a temp dir configured as an "existing" project (so `start`
// and evidence machinery skip the roadmap dependency graph) with .workflow/.
func tdRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: existing\n")
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// tdWorkflowJSON builds a minimal workflow state at the given step, optionally
// pinned to a plan contract and/or strict orchestration mode.
func tdWorkflowJSON(feature, step, planContract string, strict bool) string {
	mode := ""
	if strict {
		mode = "strict-subagents-v1"
	}
	return fmt.Sprintf(`{"feature":%q,"currentStep":%q,`+
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
		`"orchestrationMode":%q,"planContract":%q}`, feature, step, mode, planContract)
}

// tdWriteWorkflow writes tdWorkflowJSON for feature at dir.
func tdWriteWorkflow(t *testing.T, dir, feature, step, planContract string, strict bool) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"), tdWorkflowJSON(feature, step, planContract, strict))
}

// tdFeatureDocs writes the feature brief and plan file RequiredPlanInputs
// expects to find on disk (for output-actionability checks).
func tdFeatureDocs(t *testing.T, dir, feature string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "docs", "features", feature+".md"), "# "+feature+"\n")
	mustWrite(t, filepath.Join(dir, "docs", "plans", feature+".md"), "# Plan\n")
}

// tdRun runs the built binary and fails the test if it errors unexpectedly.
func tdRun(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	out, code := runCent(t, bin, dir, args...)
	if code != 0 {
		t.Fatalf("run %v exit %d: %s", args, code, out)
	}
	return out
}

// jsonArray renders a JSON string array literal.
func tdJSONArray(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ",") + "]"
}
