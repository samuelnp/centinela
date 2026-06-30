package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// trapDir plants a directory at rel so os.ReadFile against it fails with a
// non-NotExist error — used to drive gatherEvidence's read-error branches.
func trapDir(t *testing.T, rel string) {
	t.Helper()
	if err := os.MkdirAll(rel, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestCov2GatherEvidenceBriefReadError(t *testing.T) {
	chdirIntoTemp(t)
	trapDir(t, filepath.Join("docs", "features", "feat.md"))
	if _, err := gatherEvidence("feat"); err == nil {
		t.Fatal("expected a brief read error")
	}
}

func TestCov2GatherEvidencePlanReadError(t *testing.T) {
	chdirIntoTemp(t)
	trapDir(t, filepath.Join("docs", "plans", "feat.md"))
	if _, err := gatherEvidence("feat"); err == nil {
		t.Fatal("expected a plan read error")
	}
}

func TestCov2GatherEvidenceGatekeeperReadError(t *testing.T) {
	chdirIntoTemp(t)
	trapDir(t, filepath.Join(".workflow", "feat-gatekeeper.md"))
	if _, err := gatherEvidence("feat"); err == nil {
		t.Fatal("expected a gatekeeper companion read error")
	}
}

func TestCov2GatherEvidenceChangelogReadError(t *testing.T) {
	chdirIntoTemp(t)
	trapDir(t, filepath.Join(".workflow", "feat-changelog.md"))
	if _, err := gatherEvidence("feat"); err == nil {
		t.Fatal("expected a changelog companion read error")
	}
}

func TestCov2BuildPRBodyPropagatesGatherError(t *testing.T) {
	chdirIntoTemp(t)
	trapDir(t, filepath.Join("docs", "features", "feat.md"))
	if _, _, err := buildPRBody("feat"); err == nil {
		t.Fatal("buildPRBody must surface the gather error")
	}
}

func TestCov2WriteChangelogPropagatesGatherError(t *testing.T) {
	chdirIntoTemp(t)
	trapDir(t, filepath.Join("docs", "features", "feat.md"))
	if _, err := writeChangelog("feat"); err == nil {
		t.Fatal("writeChangelog must surface the gather error")
	}
}

// TestCov2RunDeliverPRCommitChangelogError: a clean tree reaches commitChangelog,
// whose compose step fails because the brief source is unreadable.
func TestCov2RunDeliverPRCommitChangelogError(t *testing.T) {
	deliverRepo(t, true)
	trapDir(t, filepath.Join("docs", "features", "feat.md"))
	stubGitDeliver(t, func(args ...string) (string, error) { return "", nil })
	err := runDeliverPR(nil, "feat")
	if err == nil || !strings.Contains(err.Error(), "compose changelog") {
		t.Fatalf("expected a compose-changelog error, got %v", err)
	}
}

// TestCov2RunDeliverPRPushFails: clean tree, no changelog change, push fails.
func TestCov2RunDeliverPRPushFails(t *testing.T) {
	deliverRepo(t, true)
	stubGitDeliver(t, func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "push" {
			return "denied", os.ErrPermission
		}
		return "", nil
	})
	err := runDeliverPR(nil, "feat")
	if err == nil || !strings.Contains(err.Error(), "git push failed") {
		t.Fatalf("expected a git push failure, got %v", err)
	}
}
