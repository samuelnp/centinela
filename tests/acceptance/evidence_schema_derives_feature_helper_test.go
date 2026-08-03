// Acceptance helpers for specs/evidence-schema-skeleton-legacy-handoff.feature.
// The binary is built ONCE into a persistent temp dir and every scenario drives
// it as an agent would: no network, no shared repo state, fixtures under
// t.TempDir only.
package acceptance_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	schemaBinOnce sync.Once
	schemaBin     string
	schemaBinErr  string
)

func buildSchemaBin(t *testing.T) string {
	t.Helper()
	schemaBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-schema-bin")
		if err != nil {
			schemaBinErr = err.Error()
			return
		}
		bin := filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			schemaBinErr = err.Error() + "\n" + string(out)
			return
		}
		schemaBin = bin
	})
	if schemaBin == "" {
		t.Fatalf("build centinela: %s", schemaBinErr)
	}
	return schemaBin
}

// schemaRun runs the binary in dir and returns stdout, stderr and exit code
// SEPARATELY: the skeleton's contract is that stdout alone is the JSON payload,
// which a merged stream cannot demonstrate.
func schemaRun(t *testing.T, dir string, args ...string) (string, string, int) {
	t.Helper()
	c := exec.Command(buildSchemaBin(t), args...)
	c.Dir = dir
	stdout, err := c.Output()
	var errb []byte
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code, errb = ee.ExitCode(), ee.Stderr
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	return string(stdout), string(errb), code
}

// schemaFields parses the printed skeleton, failing loudly if it is not JSON —
// the output is embedded verbatim in prompts, so unparseable is a defect.
func schemaFields(t *testing.T, stdout string) (feature, handoff string) {
	t.Helper()
	var got struct {
		Feature   string `json:"feature"`
		HandoffTo string `json:"handoffTo"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("skeleton is not valid JSON: %v\n%s", err, stdout)
	}
	return got.Feature, got.HandoffTo
}

// schemaRepo builds <tmp>/.worktrees/<feature> with workflow state, a brief and
// a nested package directory, mirroring how a feature worktree is laid out.
// pins is the raw contract fragment ("" = a legacy workflow with neither pin).
func schemaRepo(t *testing.T, feature, order, pins, brief, step string) string {
	t.Helper()
	base := t.TempDir()
	if r, err := filepath.EvalSymlinks(base); err == nil {
		base = r
	}
	root := filepath.Join(base, ".worktrees", feature)
	if err := os.MkdirAll(filepath.Join(root, "internal", "evidence"), 0o755); err != nil {
		t.Fatal(err)
	}
	schemaWrite(t, root, ".workflow/"+feature+".json",
		`{"feature":"`+feature+`","currentStep":"`+step+`","steps":{},`+pins+
			`"stepOrder":[`+order+`]}`)
	schemaWrite(t, root, "docs/features/"+feature+".md", brief)
	return root
}

func schemaWrite(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const (
	schemaCanonicalOrder = `"plan","code","tests","validate","docs"`
	schemaModernPins     = `"validateContract":"adversarial-v1","planContract":"unified-v1",`
	schemaSlugSlot       = "<feature-slug>"
	schemaRoleSlot       = "<successor-role>"
)
