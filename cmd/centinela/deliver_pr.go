package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/gitutil"
	"github.com/samuelnp/centinela/internal/ui"
)

// gitDeliver is overridable for tests; runs `git` in the current repo.
var gitDeliver = func(args ...string) (string, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// ghAvailable and ghCreatePR are overridable seams for tests; defaults shell
// out to the real `gh` CLI. ghCreatePR opens the PR with a composed body file.
var ghAvailable = gitutil.GitHubCLIAvailable

var ghCreatePR = func(feature, bodyPath string) (string, error) {
	out, err := exec.Command("gh", "pr", "create", "--head", feature, "--body-file", bodyPath).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// runDeliverPR commits the composed changelog line, pushes the feature branch,
// and opens a PR via `gh` with an evidence-composed body. With no origin it
// refuses (no push). When `gh` is absent it still pushes, prints honest manual
// instructions, and returns an error so the exit is non-zero — it never claims
// a PR was opened.
func runDeliverPR(_ *cobra.Command, feature string) error {
	if ok, _ := gitutil.HasOriginRemote("."); !ok {
		return fmt.Errorf("no origin remote — PR delivery unavailable for %q", feature)
	}
	if dirty, _ := gitDeliver("status", "--porcelain"); strings.TrimSpace(dirty) != "" {
		return fmt.Errorf("uncommitted changes present — commit them before `deliver --via pr` for %q", feature)
	}
	if err := commitChangelog(feature); err != nil {
		return err
	}
	if out, err := gitDeliver("push", "-u", "origin", feature); err != nil {
		return fmt.Errorf("git push failed: %s: %w", out, err)
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Pushed %q to origin.", feature)))
	return openPR(feature)
}

// commitChangelog inserts the feature's changelog line and, when it changed the
// file, commits just CHANGELOG.md so the push (and the PR) include it.
func commitChangelog(feature string) error {
	changed, err := writeChangelog(feature)
	if err != nil {
		return fmt.Errorf("compose changelog for %q: %w", feature, err)
	}
	if !changed {
		return nil
	}
	if out, err := gitDeliver("add", changelogPath); err != nil {
		return fmt.Errorf("git add %s: %s: %w", changelogPath, out, err)
	}
	if out, err := gitDeliver("commit", "-m", "docs(changelog): record "+feature); err != nil {
		return fmt.Errorf("git commit changelog: %s: %w", out, err)
	}
	return nil
}

// openPR builds the composed body file and opens the PR, honoring the gh-absent
// honest-failure contract (push already happened; never claim a PR was opened).
func openPR(feature string) error {
	if !ghAvailable() {
		fmt.Println(ui.StyleYellow.Render("`gh` not available — branch pushed, but no PR was opened."))
		fmt.Println(ui.StyleMuted.Render("Open a PR manually for branch " + feature + " against the default branch."))
		return fmt.Errorf("PR not opened: gh CLI unavailable for %q", feature)
	}
	bodyPath, err := buildPRBody(feature)
	if err != nil {
		return fmt.Errorf("compose PR body for %q: %w", feature, err)
	}
	defer os.Remove(bodyPath)
	url, err := ghCreatePR(feature, bodyPath)
	if err != nil {
		return fmt.Errorf("gh pr create failed: %s: %w", url, err)
	}
	fmt.Println(ui.RenderSuccess("Opened pull request: " + url))
	return nil
}
