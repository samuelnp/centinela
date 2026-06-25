package gitutil

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitHubCLIAvailable reports whether the `gh` CLI is on PATH. PR creation
// degrades to push + manual instructions when it is not.
func GitHubCLIAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// DeliveryDirective builds the two-line CENTINELA DIRECTIVE telling the
// orchestrator to ask the user how to deliver, listing only the offered
// options as exact commands. The shape mirrors MergeOutcome.StewardDirective:
// an imperative line plus a details line. When no option applies it collapses
// to a single line stating no delivery target was detected.
func DeliveryDirective(feature string, opts []Option) string {
	if len(opts) == 0 {
		return fmt.Sprintf(
			"CENTINELA DIRECTIVE: %q is complete but no delivery target detected — "+
				"configure an origin remote or use worktree mode, then deliver.",
			feature)
	}
	cmds := make([]string, 0, len(opts))
	for _, o := range opts {
		cmds = append(cmds, fmt.Sprintf("`centinela deliver %s --via %s`", feature, o))
	}
	return fmt.Sprintf(
		"CENTINELA DIRECTIVE: %q is complete — ask the user how to deliver it; "+
			"do NOT push or merge without their explicit choice.\n"+
			"Valid options: %s. Run only the one the user picks.",
		feature, strings.Join(cmds, ", "))
}
