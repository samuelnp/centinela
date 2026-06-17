// Acceptance: specs/precommit-and-pr-gate.feature
package acceptance_test

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func skipWin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git/hook acceptance tests skipped on windows")
	}
}

// Scenario: Staging a change that violates a fail gate blocks the commit and names the failing gate
func TestPrecommit_StagedFailBlocks(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "")
	writeFile(t, dir, "internal/oversized.go", pcLines(140))
	pcGit(t, dir, "add", "internal/oversized.go")
	out, code := runCent(t, buildCent(t), dir, "precommit")
	if code == 0 {
		t.Fatalf("oversized staged file must block, exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "G1") {
		t.Fatalf("output must name the failing G1 gate: %q", out)
	}
	if strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("must not contain a stack trace: %q", out)
	}
}

// Scenario: Staging only clean changes passes precommit and exits 0
func TestPrecommit_CleanPasses(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "")
	writeFile(t, dir, "internal/clean.go", pcLines(20))
	pcGit(t, dir, "add", "internal/clean.go")
	out, code := runCent(t, buildCent(t), dir, "precommit")
	if code != 0 {
		t.Fatalf("clean staged change must pass, exit=%d\n%s", code, out)
	}
}

// Scenario: Unstaged working-tree changes are ignored by precommit
func TestPrecommit_UnstagedIgnored(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "")
	writeFile(t, dir, "internal/clean.go", pcLines(20))
	pcGit(t, dir, "add", "internal/clean.go")
	writeFile(t, dir, "internal/unstaged.go", pcLines(140)) // NOT staged
	out, code := runCent(t, buildCent(t), dir, "precommit")
	if code != 0 {
		t.Fatalf("unstaged oversized file must not block, exit=%d\n%s", code, out)
	}
	if strings.Contains(out, "unstaged.go") {
		t.Fatalf("unstaged violation must not be reported: %q", out)
	}
}

// Scenario: Outside a git repo or with nothing staged precommit exits 0 cleanly
func TestPrecommit_NothingStagedAndNoRepo(t *testing.T) {
	skipWin(t)
	bin := buildCent(t)
	dir := pcRepo(t, "")
	out, code := runCent(t, bin, dir, "precommit") // nothing staged
	if code != 0 || strings.Contains(strings.ToLower(out), "panic") {
		t.Fatalf("nothing staged must exit 0 cleanly, exit=%d\n%s", code, out)
	}
	noRepo := t.TempDir()
	writeFile(t, noRepo, "centinela.toml", pcToml)
	out2, code2 := runCent(t, bin, noRepo, "precommit") // not a git repo
	if code2 != 0 || strings.Contains(strings.ToLower(out2), "panic") {
		t.Fatalf("outside a git repo must exit 0 cleanly, exit=%d\n%s", code2, out2)
	}
}

// Scenario: Precommit does not run the cross-compile build gate by default
func TestPrecommit_SkipsBuildGate(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "\n[gates.build]\nenabled = true\n")
	writeFile(t, dir, "internal/clean.go", pcLines(20))
	pcGit(t, dir, "add", "internal/clean.go")
	out, code := runCent(t, buildCent(t), dir, "precommit")
	if code != 0 {
		t.Fatalf("precommit must stay fast and pass, exit=%d\n%s", code, out)
	}
	if strings.Contains(strings.ToLower(out), "cross-compile") || strings.Contains(out, "Build") {
		t.Fatalf("build gate must not run under precommit: %q", out)
	}
}

// Scenario: A failing warn-severity gate under precommit is reported but does not block the commit
func TestPrecommit_WarnGateNonBlocking(t *testing.T) {
	skipWin(t)
	warn := "\n[[gates.custom]]\nenabled = true\nname = \"style-nit\"\ncommand = \"false\"\nseverity = \"warn\"\n"
	dir := pcRepo(t, warn)
	writeFile(t, dir, "internal/clean.go", pcLines(20))
	pcGit(t, dir, "add", "internal/clean.go")
	out, code := runCent(t, buildCent(t), dir, "precommit")
	if code != 0 {
		t.Fatalf("a warn gate must not block the commit, exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "style-nit") {
		t.Fatalf("warn gate must still be reported: %q", out)
	}
}

// Scenario: The installer writes an executable pre-commit hook that calls centinela precommit
func TestPrecommit_InstallWritesExecutableHook(t *testing.T) {
	skipWin(t)
	dir := pcRepo(t, "")
	_, code := runCent(t, buildCent(t), dir, "precommit", "install")
	if code != 0 {
		t.Fatalf("install must exit 0, got %d", code)
	}
	info, err := os.Stat(hookPath(dir))
	if err != nil {
		t.Fatalf("hook not created: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("hook must be executable, mode=%v", info.Mode().Perm())
	}
	body, _ := os.ReadFile(hookPath(dir))
	if !strings.Contains(string(body), "centinela precommit") {
		t.Fatalf("hook must invoke centinela precommit: %q", body)
	}
}

// Scenario: Installing the pre-commit hook twice leaves a single centinela block
func TestPrecommit_InstallTwiceSingleBlock(t *testing.T) {
	skipWin(t)
	bin := buildCent(t)
	dir := pcRepo(t, "")
	runCent(t, bin, dir, "precommit", "install")
	first, _ := os.ReadFile(hookPath(dir))
	_, code := runCent(t, bin, dir, "precommit", "install")
	if code != 0 {
		t.Fatalf("second install must exit 0, got %d", code)
	}
	second, _ := os.ReadFile(hookPath(dir))
	if string(first) != string(second) {
		t.Fatalf("re-install must be byte-identical (idempotent)")
	}
	if strings.Count(string(second), "# >>> centinela >>>") != 1 {
		t.Fatalf("exactly one centinela block expected: %q", second)
	}
}

// Scenario: The installer preserves a pre-existing pre-commit hook and uninstall removes only its own block
func TestPrecommit_InstallPreservesUninstallRemovesBlock(t *testing.T) {
	skipWin(t)
	bin := buildCent(t)
	dir := pcRepo(t, "")
	writeFile(t, dir, ".git/hooks/pre-commit", "#!/bin/sh\necho pre-existing-hook\n")
	runCent(t, bin, dir, "precommit", "install")
	body, _ := os.ReadFile(hookPath(dir))
	if !strings.Contains(string(body), "echo pre-existing-hook") || !strings.Contains(string(body), "# >>> centinela >>>") {
		t.Fatalf("install must preserve user hook and append the block: %q", body)
	}
	_, code := runCent(t, bin, dir, "precommit", "uninstall")
	if code != 0 {
		t.Fatalf("uninstall must exit 0, got %d", code)
	}
	after, _ := os.ReadFile(hookPath(dir))
	if !strings.Contains(string(after), "echo pre-existing-hook") {
		t.Fatalf("uninstall must keep the user line: %q", after)
	}
	if strings.Contains(string(after), "# >>> centinela >>>") {
		t.Fatalf("uninstall must remove the centinela block: %q", after)
	}
}
