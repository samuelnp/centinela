package migration

import "strings"

func mergeContent(current, template string) (string, int, int) {
	kept := extractKeepBlocks(current)
	merged, keepCount := replaceKeepBlocks(template, kept)
	custom := extractCustomSections(current, template)
	merged = appendCustomSections(merged, custom)
	if len(kept) > keepCount {
		merged = appendMissingKeepBlocks(merged, kept, keepCount)
		keepCount = len(kept)
	}
	return merged, keepCount, len(custom)
}

func appendMissingKeepBlocks(base string, blocks map[string]string, already int) string {
	if len(blocks) == already {
		return base
	}
	var entries []string
	for id, body := range blocks {
		if strings.Contains(base, "<!-- centinela:keep:start:"+id+" -->") {
			continue
		}
		block := "<!-- centinela:keep:start:" + id + " -->"
		if body != "" {
			block += "\n" + body
		}
		block += "\n<!-- centinela:keep:end:" + id + " -->"
		entries = append(entries, block)
	}
	if len(entries) == 0 {
		return base
	}
	s := "\n\n## Preserved Keep Blocks\n\n" + strings.Join(entries, "\n\n")
	return strings.TrimRight(base, "\n") + s + "\n"
}
