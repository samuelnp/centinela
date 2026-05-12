package docgen

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const KBDir = "docs/project-docs/kb"

func loadKBPages() ([]KBPage, error) {
	paths, _ := filepath.Glob(filepath.Join(KBDir, "*.md"))
	sort.Strings(paths)
	out := []KBPage{}
	for _, p := range paths {
		page, err := parseKBFile(p, readFile(p))
		if err != nil {
			return nil, err
		}
		out = append(out, page)
	}
	return out, nil
}

func parseKBFile(path, raw string) (KBPage, error) {
	feature := strings.TrimSuffix(filepath.Base(path), ".md")
	page := KBPage{Feature: feature}
	body := raw
	if strings.HasPrefix(raw, "---\n") {
		end := strings.Index(raw[4:], "\n---")
		if end < 0 {
			return page, fmt.Errorf("kb %s: unterminated frontmatter", feature)
		}
		readFrontmatter(raw[4:4+end], &page)
		body = strings.TrimPrefix(raw[4+end+4:], "\n")
	}
	sections := splitH2(body)
	required := map[string]*string{
		"What it does":      &page.WhatItDoes,
		"When you'd use it": &page.WhenToUse,
		"How it behaves":    &page.HowItBehaves,
	}
	for name, dest := range required {
		v, ok := sections[name]
		if !ok || strings.TrimSpace(v) == "" {
			return page, fmt.Errorf("kb %s: missing required section %q", feature, name)
		}
		*dest = strings.TrimSpace(v)
	}
	page.Examples = strings.TrimSpace(sections["Examples"])
	return page, nil
}

func readFrontmatter(block string, p *KBPage) {
	for line := range strings.SplitSeq(block, "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		switch k {
		case "summary":
			p.Summary = v
		case "audience":
			p.Audience = v
		case "status":
			p.Status = v
		}
	}
}

func splitH2(body string) map[string]string {
	out := map[string]string{}
	parts := strings.Split("\n"+body, "\n## ")
	for _, p := range parts[1:] {
		head, rest, ok := strings.Cut(p, "\n")
		if !ok {
			out[strings.TrimSpace(p)] = ""
			continue
		}
		out[strings.TrimSpace(head)] = rest
	}
	return out
}
