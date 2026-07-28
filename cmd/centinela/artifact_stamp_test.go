package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gatereport"
	"github.com/samuelnp/centinela/internal/workflow"
)

// stampRepo builds a real temp git repository with one commit and a workflow
// at the validate step, so `artifact stamp` runs against actual git output.
func stampRepo(t *testing.T, report string) {
	t.Helper()
	t.Chdir(t.TempDir())
	for _, args := range [][]string{
		{"init", "-q"}, {"config", "user.email", "t@t"}, {"config", "user.name", "t"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	os.MkdirAll(workflow.WorkflowDir, 0o755)                //nolint:errcheck
	os.WriteFile("src.go", []byte("package main\n"), 0o644) //nolint:errcheck
	exec.Command("git", "add", "src.go").Run()              //nolint:errcheck
	exec.Command("git", "commit", "-qm", "init").Run()      //nolint:errcheck
	if err := workflow.Save(workflow.New("f")); err != nil {
		t.Fatal(err)
	}
	if report != "" {
		os.WriteFile(workflow.GatekeeperReportPath("f"), []byte(report), 0o644) //nolint:errcheck
	}
}

const stampBody = "**Status:** SAFE\n\n```json centinela:verification\n" +
	`{"revision":"","treeDigest":"","commands":[{"argv":["centinela","validate"],"exitCode":0}]}` +
	"\n```\n"

func TestRunArtifactStampAgainstRealGit(t *testing.T) {
	stampRepo(t, stampBody)
	out := captureStdout(t, func() {
		if err := runArtifactStamp(nil, []string{"f"}); err != nil {
			t.Fatalf("runArtifactStamp: %v", err)
		}
	})
	if !strings.Contains(out, "revision:") || !strings.Contains(out, "sha256:") {
		t.Fatalf("stamp output missing the snapshot: %s", out)
	}
	data, _ := os.ReadFile(workflow.GatekeeperReportPath("f"))
	v, err := gatereport.ParseVerification(string(data))
	if err != nil || len(v.Revision) != 40 || !strings.HasPrefix(v.TreeDigest, "sha256:") {
		t.Fatalf("block not stamped from git: %+v %v", v, err)
	}
	if err := gatereport.Assess(string(data)); err != nil {
		t.Fatalf("stamped report inadmissible: %v", err)
	}
}

func TestRunArtifactStampCreatesBlockButStaysInadmissible(t *testing.T) {
	stampRepo(t, "**Status:** SAFE\n")
	if err := captureErr(t); err != nil {
		t.Fatal(err)
	}
}

func TestRunArtifactStampMissingReport(t *testing.T) {
	stampRepo(t, "")
	if err := runArtifactStamp(nil, []string{"f"}); err == nil {
		t.Fatal("want a missing-report error")
	}
}

func TestRunArtifactStampUnknownFeature(t *testing.T) {
	stampRepo(t, stampBody)
	if err := runArtifactStamp(nil, []string{"nope"}); err == nil {
		t.Fatal("want an unknown-feature error")
	}
}

// captureErr stamps a report that has no commands and asserts fail-closed.
func captureErr(t *testing.T) error {
	t.Helper()
	captureStdout(t, func() {
		if err := runArtifactStamp(nil, []string{"f"}); err != nil {
			t.Fatalf("runArtifactStamp: %v", err)
		}
	})
	data, _ := os.ReadFile(workflow.GatekeeperReportPath("f"))
	if err := gatereport.Assess(string(data)); err == nil {
		t.Fatal("stamping alone must not make a stub admissible")
	}
	return nil
}
