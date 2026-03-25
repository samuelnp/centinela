package autostart

import (
	"encoding/json"
	"strings"
	"time"
	"unicode"
)

func ExtractPrompt(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if text == "" || !strings.HasPrefix(text, "{") {
		return text
	}
	var data map[string]any
	if json.Unmarshal(raw, &data) != nil {
		return text
	}
	for _, key := range []string{"prompt", "input", "message"} {
		if s, ok := data[key].(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	if in, ok := data["input"].(map[string]any); ok {
		if s, ok := in["text"].(string); ok {
			return s
		}
	}
	return text
}

func ShouldStart(prompt string) bool {
	p := strings.ToLower(prompt)
	if p == "" || strings.Contains(p, "shall i advance") {
		return false
	}
	needles := []string{"i want to add", "please add", "let's add", "new feature", "extend "}
	for _, n := range needles {
		if strings.Contains(p, n) {
			return true
		}
	}
	return false
}

func DeriveFeature(prompt string) string {
	base := strings.ToLower(prompt)
	repl := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, base)
	stop := map[string]bool{"i": true, "want": true, "to": true, "please": true, "lets": true, "add": true, "new": true, "feature": true, "extend": true}
	parts := strings.Fields(repl)
	out := make([]string, 0, 6)
	for _, p := range parts {
		if len(out) == 6 {
			break
		}
		if len(p) < 3 || stop[p] {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return "feature-" + time.Now().UTC().Format("20060102-1504")
	}
	name := strings.Join(out, "-")
	if name[0] >= '0' && name[0] <= '9' {
		return "feature-" + name
	}
	return name
}
