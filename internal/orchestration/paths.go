package orchestration

import "fmt"

func MarkdownPath(feature string, role Role) string {
	return fmt.Sprintf(".workflow/%s-%s.md", feature, role)
}

func JSONPath(feature string, role Role) string {
	return fmt.Sprintf(".workflow/%s-%s.json", feature, role)
}
