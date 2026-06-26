package main

import (
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/delivery"
	"github.com/samuelnp/centinela/internal/evidence"
)

// changelogPath is the repo-root changelog the delivery flow appends to.
const changelogPath = "CHANGELOG.md"

// readOptional returns a file's contents, or "" when it does not exist. Any
// other read error is surfaced so genuine I/O faults are not hidden.
func readOptional(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// gatherEvidence reads every delivery source for feature from disk. A missing
// source becomes an empty section (graceful degradation), never an error.
func gatherEvidence(feature string) (delivery.Evidence, error) {
	brief, err := readOptional(filepath.Join("docs", "features", feature+".md"))
	if err != nil {
		return delivery.Evidence{}, err
	}
	plan, err := readOptional(filepath.Join("docs", "plans", feature+".md"))
	if err != nil {
		return delivery.Evidence{}, err
	}
	gate, err := evidence.ReadCompanion(feature, evidence.Role("gatekeeper"))
	if err != nil {
		return delivery.Evidence{}, err
	}
	stub, err := evidence.ReadCompanion(feature, evidence.Role("changelog"))
	if err != nil {
		return delivery.Evidence{}, err
	}
	specPath := filepath.Join("specs", feature+".feature")
	if _, statErr := os.Stat(specPath); statErr != nil {
		specPath = ""
	}
	return delivery.Evidence{
		Feature: feature, Brief: brief, Plan: plan,
		GatekeeperReport: gate, ChangelogStub: stub, SpecPath: specPath,
	}, nil
}

// buildPRBody composes the PR title and body from feature evidence, writes the
// body to a temp file, and returns (title, bodyFilePath). `gh pr create` needs
// both --title and --body-file when non-interactive.
func buildPRBody(feature string) (title, path string, err error) {
	e, err := gatherEvidence(feature)
	if err != nil {
		return "", "", err
	}
	f, err := os.CreateTemp("", "centinela-pr-body-*.md")
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	if _, err := f.WriteString(delivery.ComposePRBody(e)); err != nil {
		return "", "", err
	}
	return delivery.ComposePRTitle(e), f.Name(), nil
}

// writeChangelog inserts the feature's changelog line into CHANGELOG.md,
// idempotently. It returns whether the file changed. A missing CHANGELOG.md is
// a no-op (degradation), not an error.
func writeChangelog(feature string) (bool, error) {
	e, err := gatherEvidence(feature)
	if err != nil {
		return false, err
	}
	current, err := readOptional(changelogPath)
	if err != nil || current == "" {
		return false, err
	}
	updated, changed := delivery.InsertEntry(current, delivery.ComposeChangelog(e))
	if !changed {
		return false, nil
	}
	return true, os.WriteFile(changelogPath, []byte(updated), 0o644)
}
