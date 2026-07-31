package workflow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func validateDocsOutput(feature string) error {
	if feature == "" {
		return fmt.Errorf("feature is required for docs validation")
	}
	return validateChangelog(feature)
}

// validateChangelog requires a real changelog entry for every feature: at
// least one non-blank line, and no line still carrying a slot the scaffold
// `centinela artifact new <feature> changelog` wrote. A stub is non-empty, so
// "non-blank" alone let the docs step complete on an entry nobody wrote.
//
// The check is CONTENT-scoped, not positional. Inspecting only the first
// non-blank line made it cosmetic: "## Changelog" in front of an untouched
// stub passed, and adding a heading to a report is one of the most natural
// things an author does. Every line is now asked the same question, so a
// half-filled entry fails with the rest. Prose that legitimately quotes the
// marker still passes — the generic "<FILL: ...>" citation form is not a slot
// (see orchestration.UnreplacedSlot).
//
// The real-updated-doc-file rule for user-facing features lives in
// orchestration evidence validation (documentation-specialist outputs), not
// here.
func validateChangelog(feature string) error {
	path := filepath.Join(WorkflowDir, feature+"-changelog.md")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("changelog entry missing for %q: %s (write a one-line summary, e.g. via: centinela artifact new %s changelog)", feature, path, feature)
	}
	defer f.Close() //nolint:errcheck // read-only handle
	scanner := bufio.NewScanner(f)
	entries, lineNo := 0, 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		entries++
		if orchestration.UnreplacedSlot(line) {
			return fmt.Errorf("changelog entry is still a template placeholder for %q: %s line %d (replace the <FILL: ...> slots with a real one-line summary of the change; the generic \"<FILL: ...>\" citation form is not a slot, so prose may still quote the marker)", feature, path, lineNo)
		}
	}
	if entries == 0 {
		return fmt.Errorf("changelog entry is empty for %q: %s (write a one-line summary of the change)", feature, path)
	}
	return nil
}
