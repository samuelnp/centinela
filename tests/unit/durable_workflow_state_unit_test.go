package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// dwsRepo chdirs into an empty project with a .workflow/ directory.
func dwsRepo(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	t.Chdir(d)
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

func dwsWrite(t *testing.T, feature, body string) {
	t.Helper()
	if err := os.WriteFile(workflow.FilePath(feature), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The version rules, stated through the public API a caller actually uses.
func TestSchemaVersionRules(t *testing.T) {
	cases := []struct {
		name, body  string
		wantVersion int
		saveRefused bool
	}{
		{"absent means 1", `{"feature":"f","currentStep":"code","steps":{}}`, 1, false},
		{"null means 1", `{"schemaVersion":null,"feature":"f","currentStep":"code","steps":{}}`, 1, false},
		{"zero means 1", `{"schemaVersion":0,"feature":"f","currentStep":"code","steps":{}}`, 1, false},
		{"equal round-trips", `{"schemaVersion":1,"feature":"f","currentStep":"code","steps":{}}`, 1, false},
		{"higher is refused", `{"schemaVersion":99,"feature":"f","currentStep":"code","steps":{}}`, 99, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dwsRepo(t)
			dwsWrite(t, "f", tc.body)
			wf, err := workflow.Load("f")
			if err != nil {
				t.Fatalf("every version must load: %v", err)
			}
			if wf.SchemaVersion != tc.wantVersion {
				t.Fatalf("SchemaVersion = %d, want %d", wf.SchemaVersion, tc.wantVersion)
			}
			err = workflow.Save(wf)
			if tc.saveRefused != (err != nil) {
				t.Fatalf("save refused = %v, want %v (err: %v)", err != nil, tc.saveRefused, err)
			}
			if tc.saveRefused && string(mustRead(t, workflow.FilePath("f"))) != tc.body {
				t.Fatal("a refused save must leave the file byte-identical")
			}
		})
	}
}

// A newly started workflow is stamped, and the key is first so `head -3` of any
// state file shows it.
func TestSaveStampsVersionFirst(t *testing.T) {
	dwsRepo(t)
	if err := workflow.Save(workflow.New("beta")); err != nil {
		t.Fatal(err)
	}
	raw := string(mustRead(t, workflow.FilePath("beta")))
	if !strings.Contains(raw, `"schemaVersion": 1`) {
		t.Fatalf("state file is not stamped: %s", raw)
	}
	if idx := strings.Index(raw, `"schemaVersion"`); idx > 5 {
		t.Fatalf("schemaVersion must be the first key, found at %d: %s", idx, raw)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
