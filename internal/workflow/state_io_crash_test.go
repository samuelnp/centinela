package workflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const crashEnv = "CENTINELA_CRASH_TEST"

// TestHelperCrashesBeforeRename is not a test: it is the child half of
// TestKilledWriteLeavesPreviousStateIntact. It runs only when re-exec'd with
// crashEnv set, and dies via os.Exit — which skips deferred cleanup exactly the
// way a SIGKILL does — in the window between the fsynced temp write and the
// rename that publishes it.
func TestHelperCrashesBeforeRename(t *testing.T) {
	if os.Getenv(crashEnv) != "1" {
		t.Skip("helper process; driven by TestKilledWriteLeavesPreviousStateIntact")
	}
	t.Chdir(os.Getenv(crashEnv + "_DIR"))
	beforeRename = func() { os.Exit(1) }
	wf, err := Load("alpha")
	if err != nil {
		os.Exit(2)
	}
	wf.CurrentStep = "tests"
	_ = Save(wf)
	os.Exit(3) // reached only if the seam failed to fire
}

// TestKilledWriteLeavesPreviousStateIntact is the real proof for "a killed
// write leaves the previous state intact". Racing a signal against rename(2)
// is nondeterministic, so the kill is made deterministic with the beforeRename
// seam in a re-exec'd child.
func TestKilledWriteLeavesPreviousStateIntact(t *testing.T) {
	dir := stateRepo(t)
	wf := New("alpha")
	wf.CurrentStep = "code"
	if err := Save(wf); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(FilePath("alpha"))
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "^TestHelperCrashesBeforeRename$")
	cmd.Env = append(os.Environ(), crashEnv+"=1", crashEnv+"_DIR="+dir)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("helper was supposed to die before the rename: %s", out)
	}
	if strings.Contains(string(out), "exit status 3") {
		t.Fatalf("the beforeRename seam never fired: %s", out)
	}

	after, err := os.ReadFile(FilePath("alpha"))
	if err != nil {
		t.Fatalf("the state file must survive a killed write: %v", err)
	}
	if len(after) == 0 {
		t.Fatal("a zero-byte state file was left on disk")
	}
	reloaded, err := Load("alpha")
	if err != nil {
		t.Fatalf("the state file must still parse: %v", err)
	}
	if reloaded.CurrentStep != "code" {
		t.Fatalf("step = %q, want the pre-crash %q", reloaded.CurrentStep, "code")
	}
	if string(after) != string(before) {
		t.Fatal("the target was modified by a write that never reached its rename")
	}
}

// The leftover temp a SIGKILL strands must not trip the doctor evidence check,
// whose repair cannot remove a non-role temp and would stay red forever.
func TestKilledWriteTempIsNotAnEvidenceOrphan(t *testing.T) {
	stateRepo(t)
	matches, err := filepath.Glob(filepath.Join(WorkflowDir, "*.json.tmp"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("precondition: %v %v", matches, err)
	}
	tmp, err := writeTempSibling(WorkflowDir, "alpha.json", []byte("{}"), stateFileMode)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(tmp) })
	matches, _ = filepath.Glob(filepath.Join(WorkflowDir, "*.json.tmp"))
	if len(matches) != 0 {
		t.Fatalf("doctor would report %v as orphaned evidence", matches)
	}
}
