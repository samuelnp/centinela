package treestate

import (
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes a git subcommand in dir and returns its stdout. It is the
// single seam this package touches the outside world through, so callers can
// stub the tree state without a real repository.
type Runner func(dir string, args ...string) (string, error)

// NewExecRunner returns the production Runner. stderr is folded into the error
// so a git failure names its own cause instead of a bare exit status.
func NewExecRunner() Runner {
	return func(dir string, args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
		}
		return string(out), nil
	}
}
