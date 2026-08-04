package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/doctor"
	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/workflow"
)

// dwsBareProject creates and enters an empty project dir with a .workflow/.
func dwsBareProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	t.Chdir(dir)
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// dwsProject is a bare project plus one active workflow.
func dwsProject(t *testing.T) string {
	t.Helper()
	dir := dwsBareProject(t)
	if err := workflow.Save(workflow.New("alpha")); err != nil {
		t.Fatal(err)
	}
	return dir
}

// dwsDiagnose runs the real doctor over dir and returns its diagnoses by name.
func dwsDiagnose(t *testing.T, dir string) map[string]doctor.Diagnosis {
	t.Helper()
	ctx, err := doctor.NewContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]doctor.Diagnosis{}
	for _, d := range doctor.Run(ctx) {
		out[d.Name] = d
	}
	return out
}

// Scenario: An abandoned temporary file is not mistaken for orphaned evidence.
//
// Asserted on doctor's own diagnosis rather than on filepath.Glob: the glob is
// the mechanism, the check's verdict is the promise. A temp that tripped this
// check would make `doctor` red forever, since evidence.Repair only knows how
// to remove <feature>-<role> temps and would report a fix that removes nothing.
func TestDoctorStaysGreenWithAWorkflowTemp(t *testing.T) {
	dir := dwsProject(t)
	tmp := filepath.Join(workflow.WorkflowDir, ".alpha.json.tmp-1234567890")
	if err := os.WriteFile(tmp, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := dwsDiagnose(t, dir)
	ev, ok := got["evidence"]
	if !ok {
		t.Fatal("doctor ran no evidence check")
	}
	if ev.Status != doctor.OK || ev.Message != "no orphaned evidence temp files" {
		t.Fatalf("a workflow temp must not be seen as orphaned evidence: %+v", ev)
	}
	for name, d := range got {
		for _, detail := range d.Details {
			if strings.Contains(detail, ".alpha.json.tmp-") {
				t.Fatalf("check %q reported the workflow temp: %s", name, detail)
			}
		}
	}

	// A live write's temp must survive a repair sweep — reaping it mid-flight
	// would corrupt the very write this feature made atomic.
	if _, err := evidence.Repair("alpha"); err != nil {
		t.Fatalf("evidence repair: %v", err)
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Fatalf("evidence.Repair removed a workflow temp: %v", err)
	}
}

// dwsWriteRaw drops hand-authored bytes at .workflow/<feature>.json.
func dwsWriteRaw(t *testing.T, feature, body string) {
	t.Helper()
	if err := os.WriteFile(workflow.FilePath(feature), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
