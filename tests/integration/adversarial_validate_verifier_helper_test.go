// Integration coverage for docs/plans/adversarial-validate-verifier.md.
// See adversarial_validate_verifier_integration_test.go for the scenario.
package integration_test

import (
	"os"
	"os/exec"
	"testing"
)

func avviGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func avviWorkflow(t *testing.T, dir, feature string) {
	t.Helper()
	mustWrite(t, dir, ".workflow/"+feature+".json",
		`{"feature":"`+feature+`","currentStep":"validate",`+
			`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
			`"validateContract":"adversarial-v1"}`)
}

// avviWriteVerification writes a minimal SAFE report with the given stamp.
func avviWriteVerification(t *testing.T, path, revision, digest string) {
	t.Helper()
	body := "**Status:** SAFE\n\n```json centinela:verification\n" +
		`{"revision":"` + revision + `","treeDigest":"` + digest +
		`","commands":[{"argv":["centinela","validate"],"exitCode":0}]}` + "\n```\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func avviChdir(t *testing.T, dir string, fn func()) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old) //nolint:errcheck
	fn()
}
