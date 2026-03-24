package projectstage

import (
	"os"
	"strings"
)

const (
	Greenfield = "greenfield"
	Existing   = "existing"
)

func Load(projectFile string) (string, error) {
	data, err := os.ReadFile(projectFile)
	if err != nil {
		return "", err
	}
	return Parse(string(data)), nil
}

func Parse(markdown string) string {
	for _, raw := range strings.Split(markdown, "\n") {
		line := strings.ToLower(strings.TrimSpace(raw))
		if !strings.Contains(line, "project stage") || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		value := sanitize(parts[1])
		switch value {
		case Existing:
			return Existing
		case Greenfield:
			return Greenfield
		}
	}
	return Greenfield
}

func sanitize(value string) string {
	v := strings.TrimSpace(value)
	v = strings.Trim(v, "*`<>_")
	fields := strings.Fields(v)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
