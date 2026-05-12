package docgen

import (
	"fmt"
	"sort"
	"strings"
)

func RenderKBIndex(d *Data) string {
	b := &strings.Builder{}
	fmt.Fprint(b, "<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>Knowledge Base — ")
	fmt.Fprint(b, esc(d.Title), "</title>", renderHeadAssets(), "</head><body><div class=\"page\">")
	fmt.Fprint(b, kbSidebar(d), "<main class=\"content\">")
	fmt.Fprint(b, kbHero(len(d.KB), len(kbAllFeatures(d))), kbCards(d))
	fmt.Fprint(b, "</main></div></body></html>")
	return b.String()
}

func RenderKBFeature(p KBPage, title string) string {
	b := &strings.Builder{}
	fmt.Fprint(b, "<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>")
	fmt.Fprint(b, esc(p.Feature), " — ", esc(title), "</title>", renderHeadAssets(), "</head><body><div class=\"page\">")
	fmt.Fprint(b, kbFeatureSidebar(p), "<main class=\"content\">", kbFeatureHero(p))
	fmt.Fprint(b, kbSection("What it does", p.WhatItDoes))
	fmt.Fprint(b, kbSection("When you'd use it", p.WhenToUse))
	fmt.Fprint(b, kbSection("How it behaves", p.HowItBehaves))
	if strings.TrimSpace(p.Examples) != "" {
		fmt.Fprint(b, kbSection("Examples", p.Examples))
	}
	fmt.Fprint(b, "</main></div></body></html>")
	return b.String()
}

func kbAllFeatures(d *Data) []string {
	seen := map[string]bool{}
	for _, s := range d.Specs {
		name := strings.TrimSuffix(specBase(s), ".feature")
		seen[name] = true
	}
	for _, f := range d.FeatureDocs {
		name := strings.TrimSuffix(specBase(f), ".md")
		seen[name] = true
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func specBase(path string) string {
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return path
	}
	return path[i+1:]
}
