package migration

import "strings"

func extractKeepBlocks(content string) map[string]string {
	out := map[string]string{}
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		id, ok := keepStartID(lines[i])
		if !ok {
			continue
		}
		j := i + 1
		for ; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "<!-- centinela:keep:end:"+id+" -->" {
				break
			}
		}
		if j < len(lines) {
			out[id] = strings.Join(lines[i+1:j], "\n")
			i = j
		}
	}
	return out
}

func keepStartID(line string) (string, bool) {
	trim := strings.TrimSpace(line)
	const p = "<!-- centinela:keep:start:"
	if !strings.HasPrefix(trim, p) || !strings.HasSuffix(trim, " -->") {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(trim, p), " -->")
	return id, id != ""
}

func replaceKeepBlocks(template string, blocks map[string]string) (string, int) {
	if len(blocks) == 0 {
		return template, 0
	}
	lines := strings.Split(template, "\n")
	out := make([]string, 0, len(lines))
	preserved := 0
	for i := 0; i < len(lines); i++ {
		id, ok := keepStartID(lines[i])
		if !ok {
			out = append(out, lines[i])
			continue
		}
		body, exists := blocks[id]
		end := "<!-- centinela:keep:end:" + id + " -->"
		j := i + 1
		for ; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == end {
				break
			}
		}
		out = append(out, lines[i])
		if exists {
			if body != "" {
				out = append(out, strings.Split(body, "\n")...)
			}
			preserved++
		} else if j > i+1 {
			out = append(out, lines[i+1:j]...)
		}
		out = append(out, end)
		if j < len(lines) {
			i = j
		}
	}
	return strings.Join(out, "\n"), preserved
}
