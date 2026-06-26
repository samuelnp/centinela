package main

import (
	"strings"
	"testing"
)

// TestPRCreateArgsIncludeTitle guards the dogfood regression: dropping --fill
// for --body-file means `gh pr create` now REQUIRES an explicit --title, or it
// errors non-interactively. The argv must carry --title and --body-file and
// never --fill.
func TestPRCreateArgsIncludeTitle(t *testing.T) {
	args := prCreateArgs("feat", "feat: do the thing", "/tmp/body.md")
	joined := strings.Join(args, " ")
	for _, want := range []string{"--head feat", "--title feat: do the thing", "--body-file /tmp/body.md"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("prCreateArgs missing %q in: %s", want, joined)
		}
	}
	if strings.Contains(joined, "--fill") {
		t.Fatalf("prCreateArgs must not use --fill: %s", joined)
	}
}

// TestPRCreateArgsTitleAndBodyArePaired ensures --title and --body-file each
// have a following value (not a trailing bare flag).
func TestPRCreateArgsTitleAndBodyArePaired(t *testing.T) {
	args := prCreateArgs("f", "t", "b")
	idxOf := func(s string) int {
		for i, a := range args {
			if a == s {
				return i
			}
		}
		return -1
	}
	for _, flag := range []string{"--title", "--body-file"} {
		i := idxOf(flag)
		if i < 0 || i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
			t.Fatalf("%s must be followed by a value in %v", flag, args)
		}
	}
}
