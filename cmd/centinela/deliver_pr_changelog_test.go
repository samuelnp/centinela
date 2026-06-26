package main

import (
	"os"
	"strings"
	"testing"
)

// seedChangelogSources drops the changelog stub + a CHANGELOG.md with an
// [Unreleased]/Added block into the CWD (call after deliverRepo chdir).
func seedChangelogSources(t *testing.T, feature string) {
	t.Helper()
	if err := os.WriteFile(".workflow/"+feature+"-changelog.md", []byte("- feat: ship "+feature+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cl := "# Changelog\n\n## [Unreleased]\n\n### Added\n\n### Changed\n\n### Fixed\n"
	if err := os.WriteFile(changelogPath, []byte(cl), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRunDeliverPRPassesBodyFile: ghCreatePR must receive a NON-EMPTY body-file
// path (not --fill), proving the composed body is wired through.
func TestRunDeliverPRPassesBodyFile(t *testing.T) {
	deliverRepo(t, true)
	seedChangelogSources(t, "feat")
	cleanPushStub(t)
	var gotPath string
	pa, pc := ghAvailable, ghCreatePR
	ghAvailable = func() bool { return true }
	ghCreatePR = func(_, bodyPath string) (string, error) { gotPath = bodyPath; return "https://x/pull/1", nil }
	t.Cleanup(func() { ghAvailable, ghCreatePR = pa, pc })
	if err := runDeliverPR(nil, "feat"); err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if strings.TrimSpace(gotPath) == "" {
		t.Fatal("ghCreatePR must receive a non-empty --body-file path")
	}
}

// TestCommitChangelogCommitsOnlyWhenChanged: real git is used; the first call
// commits CHANGELOG.md, the second (idempotent) makes no commit.
func TestCommitChangelogCommitsOnlyWhenChanged(t *testing.T) {
	deliverRepo(t, true)
	seedChangelogSources(t, "feat")
	var adds []string
	stubGitDeliver(t, func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "add" {
			adds = append(adds, args[len(args)-1])
		}
		return "", nil
	})
	if err := commitChangelog("feat"); err != nil {
		t.Fatalf("first commitChangelog: %v", err)
	}
	if len(adds) != 1 || adds[0] != changelogPath {
		t.Fatalf("first call should add CHANGELOG.md once, got %v", adds)
	}
	if err := commitChangelog("feat"); err != nil {
		t.Fatalf("second commitChangelog: %v", err)
	}
	if len(adds) != 1 {
		t.Fatalf("idempotent re-run must not commit again, got adds=%v", adds)
	}
}
