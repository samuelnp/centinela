package treestate

import (
	"errors"
	"strings"
	"testing"
)

// stubRunner answers by subcommand so a test states only what it cares about.
func stubRunner(out map[string]string, fail string) Runner {
	return func(_ string, args ...string) (string, error) {
		if args[0] == fail {
			return "", errors.New("boom")
		}
		return out[args[0]], nil
	}
}

func TestStampReadsRevisionAndDigest(t *testing.T) {
	run := stubRunner(map[string]string{
		"rev-parse": "9f2c1ab\n",
		"status":    " M internal/x.go\n",
		"diff":      srcDiff,
	}, "")
	snap, err := Stamp(".", run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Revision != "9f2c1ab" {
		t.Fatalf("revision = %q", snap.Revision)
	}
	if snap.Digest != Digest(" M internal/x.go\n", srcDiff) {
		t.Fatalf("digest = %q", snap.Digest)
	}
}

func TestStampWorkflowOnlyDirtEqualsClean(t *testing.T) {
	clean, err := Stamp(".", stubRunner(map[string]string{"rev-parse": "abc\n"}, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dirty, err := Stamp(".", stubRunner(map[string]string{
		"rev-parse": "abc\n",
		"status":    " M .workflow/f-gatekeeper.md\n",
		"diff":      wfDiff,
	}, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clean != dirty {
		t.Fatalf("verifier's own writes staled its stamp: %+v vs %+v", clean, dirty)
	}
}

func TestStampSurfacesGitFailures(t *testing.T) {
	for _, sub := range []string{"rev-parse", "status", "diff"} {
		_, err := Stamp(".", stubRunner(nil, sub))
		if err == nil || !strings.Contains(err.Error(), "boom") {
			t.Fatalf("%s failure must surface, got %v", sub, err)
		}
	}
}

func TestNewExecRunnerReportsMissingRepo(t *testing.T) {
	if _, err := NewExecRunner()(t.TempDir(), "rev-parse", "HEAD"); err == nil {
		t.Fatal("want an error outside a repository")
	}
}
