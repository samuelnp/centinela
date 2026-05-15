package worktree_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestValidateFeatureSlug_Accepts(t *testing.T) {
	good := []string{"alpha", "alpha-beta", "feature-1", "a1b2-c3"}
	for _, s := range good {
		if err := worktree.ValidateFeatureSlug(s); err != nil {
			t.Fatalf("ValidateFeatureSlug(%q) unexpected err: %v", s, err)
		}
	}
}

func TestValidateFeatureSlug_RejectsUnsafe(t *testing.T) {
	bad := []string{
		"",                 // empty
		"Alpha",            // uppercase
		"alpha_beta",       // underscore
		"alpha beta",       // space
		"alpha/beta",       // path separator
		"alpha/../beta",    // path traversal
		"alpha;rm -rf /",   // shell injection
		"-alpha",           // leading hyphen
		"alpha-",           // trailing hyphen
		"alpha--beta",      // double hyphen
		".alpha",           // dotfile
		"αlpha",            // non-ASCII
	}
	for _, s := range bad {
		if err := worktree.ValidateFeatureSlug(s); err == nil {
			t.Fatalf("ValidateFeatureSlug(%q) unexpectedly accepted", s)
		}
	}
}
