package migration

import "testing"

func TestExtractCustomSectionsSkipsPreservedBuckets(t *testing.T) {
	current := "## A\na\n\n## Preserved Custom Sections\nx\n\n## Preserved Keep Blocks\ny\n"
	tpl := "## A\na\n"
	sections := extractCustomSections(current, tpl)
	if len(sections) != 0 {
		t.Fatal("expected preserved buckets to be skipped")
	}
}
