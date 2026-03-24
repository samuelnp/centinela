package migration

import "strings"

type Header struct {
	Version  string
	Template string
}

func ParseHeader(content string) (Header, bool) {
	line, _, _ := strings.Cut(content, "\n")
	line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
	if !strings.HasPrefix(line, "<!-- centinela:doc-version=") || !strings.HasSuffix(line, "-->") {
		return Header{}, false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(line, "<!-- "), " -->")
	toks := strings.Fields(body)
	if len(toks) < 2 || !strings.HasPrefix(toks[0], "centinela:doc-version=") {
		return Header{}, false
	}
	v := strings.TrimPrefix(toks[0], "centinela:doc-version=")
	t := strings.TrimPrefix(toks[1], "template=")
	if v == "" || t == "" {
		return Header{}, false
	}
	return Header{Version: v, Template: t}, true
}

func WithHeader(content, template, version string) string {
	if _, ok := ParseHeader(content); ok {
		_, rest, found := strings.Cut(content, "\n")
		if found {
			content = rest
		} else {
			content = ""
		}
	}
	h := "<!-- centinela:doc-version=" + version + " template=" + template + " -->"
	content = strings.TrimPrefix(content, "\n")
	if content == "" {
		return h + "\n"
	}
	return h + "\n" + content
}
