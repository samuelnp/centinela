package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestCommitStepNoPanic(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	exec.Command("git", "init").Run()        //nolint:errcheck
	os.Setenv("GIT_AUTHOR_NAME", "t")        //nolint:errcheck
	os.Setenv("GIT_AUTHOR_EMAIL", "t@t")     //nolint:errcheck
	os.Setenv("GIT_COMMITTER_NAME", "t")     //nolint:errcheck
	os.Setenv("GIT_COMMITTER_EMAIL", "t@t")  //nolint:errcheck
	os.WriteFile("a.txt", []byte("x"), 0644) //nolint:errcheck
	commitStep("f", "code", 2, 5)
	out, _ := exec.Command("git", "log", "--oneline", "-1").CombinedOutput()
	if len(out) == 0 {
		t.Fatal("expected at least one git commit")
	}
}
