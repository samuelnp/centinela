package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// commitStep stages all changes and commits with a workflow step message.
// Fails silently if git is unavailable, not a repo, or nothing to commit.
func commitStep(feature, step string, stepNum, total int) {
	add := exec.Command("git", "add", "-A")
	if err := add.Run(); err != nil {
		return // not a git repo or git not installed
	}

	msg := fmt.Sprintf(
		"chore(workflow): %s — %s complete [%d/%d]",
		feature, step, stepNum, total,
	)
	commit := exec.Command("git", "commit", "-m", msg)
	out, err := commit.CombinedOutput()
	if err != nil && !strings.Contains(string(out), "nothing to commit") {
		fmt.Fprintf(os.Stderr, "[centinela] git commit skipped: %s\n", strings.TrimSpace(string(out)))
	}
}
