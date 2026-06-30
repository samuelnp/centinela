package migration

import "testing"

func TestAppendMissingKeepBlocksAlreadyEqual(t *testing.T) {
	// len(blocks) == already: nothing is missing, base is returned untouched.
	base := "# Doc\n"
	out := appendMissingKeepBlocks(base, map[string]string{"a": "x"}, 1)
	if out != base {
		t.Fatalf("expected unchanged base, got %q", out)
	}
}

func TestAppendMissingKeepBlocksAllPresent(t *testing.T) {
	// blocks count differs from already, but every id is already present in base,
	// so the per-id continue empties entries and base is returned unchanged.
	base := "<!-- centinela:keep:start:a -->\nx\n<!-- centinela:keep:end:a -->\n"
	out := appendMissingKeepBlocks(base, map[string]string{"a": "x"}, 0)
	if out != base {
		t.Fatalf("expected base unchanged when all blocks present, got %q", out)
	}
}
