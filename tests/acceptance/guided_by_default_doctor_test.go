// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"os/exec"
	"strings"
	"testing"
)

// gbdWorkflowSeed marks doctorRepo's project as having an active workflow, so
// the profile-default check has something to consider inheriting. A matching
// git branch keeps the UNRELATED workflow-state check from also reporting
// this feature as orphaned, which would otherwise mask the exit code this
// test cares about.
func gbdWorkflowSeed(t *testing.T, dir string) {
	t.Helper()
	gitInit(t, dir)
	writeFile(t, dir, ".workflow/some-feature.json",
		`{"feature":"some-feature","currentStep":"plan","steps":{}}`)
	c := exec.Command("git", "branch", "some-feature")
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git branch: %v\n%s", err, out)
	}
}

// Scenario: Projects inheriting the new default are told about it
func TestGBD_DoctorAdvisesInheritedDefault(t *testing.T) {
	dir := doctorRepo(t)
	gbdWorkflowSeed(t, dir)
	out, code := runDoctor(t, dir)
	if code != 0 {
		t.Fatalf("the advisory must not fail doctor, got %d: %s", code, out)
	}
	if !strings.Contains(out, "default is now guided") {
		t.Fatalf("doctor must advise that the default is now guided: %s", out)
	}
}

// Scenario: A project that already pins its profile is not advised
func TestGBD_DoctorSilentWhenProfilePinned(t *testing.T) {
	dir := doctorRepo(t)
	gbdWorkflowSeed(t, dir)
	writeFile(t, dir, "centinela.toml", "[workflow]\nenforcement_profile = \"strict\"\n")
	out, code := runDoctor(t, dir)
	if code != 0 {
		t.Fatalf("doctor must exit 0, got %d: %s", code, out)
	}
	if strings.Contains(out, "default is now guided") {
		t.Fatalf("a project with an explicit pin must not be advised: %s", out)
	}
}
