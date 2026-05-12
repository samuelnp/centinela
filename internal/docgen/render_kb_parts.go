package docgen

import (
	"fmt"
	"strings"
)

func kbSidebar(d *Data) string {
	return fmt.Sprintf(`<aside class="sidebar"><div class="brand">Knowledge Base</div><div class="caption">Plain-language feature guides</div><nav class="toc"><a href="../index.html">← Back to project docs</a><a href="#overview">Overview</a><a href="#features">Features</a></nav><p class="caption">%d guides · %d features tracked</p></aside>`, len(d.KB), len(kbAllFeatures(d)))
}

func kbFeatureSidebar(p KBPage) string {
	status := p.Status
	if status == "" {
		status = "unknown"
	}
	return fmt.Sprintf(`<aside class="sidebar"><div class="brand">%s</div><div class="caption">Status: %s</div><nav class="toc"><a href="index.html">← Back to knowledge base</a><a href="#what">What it does</a><a href="#when">When you'd use it</a><a href="#how">How it behaves</a><a href="#examples">Examples</a></nav></aside>`, esc(p.Feature), esc(status))
}

func kbHero(populated, total int) string {
	return fmt.Sprintf(`<header id="overview" class="hero"><h1>Knowledge Base</h1><p class="meta">End-user guides for each Centinela feature. %d of %d features have a written guide.</p><span class="pill">Plain language</span><span class="pill">No Gherkin</span></header>`, populated, total)
}

func kbFeatureHero(p KBPage) string {
	summary := p.Summary
	if summary == "" {
		summary = "End-user guide for this feature."
	}
	return fmt.Sprintf(`<header class="hero"><h1>%s</h1><p class="meta">%s</p></header>`, esc(p.Feature), esc(summary))
}

func kbCards(d *Data) string {
	byFeature := map[string]KBPage{}
	for _, p := range d.KB {
		byFeature[p.Feature] = p
	}
	statusByFeature := map[string]string{}
	for _, s := range d.States {
		statusByFeature[s.Feature] = s.Status
	}
	b := &strings.Builder{}
	b.WriteString(`<section id="features"><h2>Feature Guides</h2><div class="cards">`)
	for _, name := range kbAllFeatures(d) {
		page, written := byFeature[name]
		st := statusByFeature[name]
		b.WriteString(kbCard(name, page, written, st))
	}
	b.WriteString("</div></section>")
	return b.String()
}

func kbCard(name string, p KBPage, written bool, status string) string {
	if !written {
		return fmt.Sprintf(`<div class="card"><div class="k">%s</div><div class="v">Guide not yet written</div><div class="caption">Status: %s</div></div>`, esc(name), esc(status))
	}
	summary := p.Summary
	if summary == "" {
		summary = "Read the guide for plain-language details."
	}
	return fmt.Sprintf(`<a class="card" href="%s.html"><div class="k">%s</div><div class="v" style="font-size:1rem;font-weight:600;line-height:1.3">%s</div><div class="caption">Status: %s</div></a>`, esc(name), esc(name), esc(summary), esc(status))
}

func kbSection(title, body string) string {
	anchor := strings.ToLower(strings.ReplaceAll(strings.Fields(title)[0], "'", ""))
	return fmt.Sprintf(`<section id="%s"><h2>%s</h2>%s</section>`, anchor, esc(title), mdToHTML(body))
}

func mdToHTML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	b := &strings.Builder{}
	inList := false
	for line := range strings.SplitSeq(s, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "- ") {
			if !inList {
				b.WriteString("<ul>")
				inList = true
			}
			b.WriteString("<li>" + esc(strings.TrimPrefix(t, "- ")) + "</li>")
			continue
		}
		if inList {
			b.WriteString("</ul>")
			inList = false
		}
		if t == "" {
			continue
		}
		b.WriteString("<p>" + esc(t) + "</p>")
	}
	if inList {
		b.WriteString("</ul>")
	}
	return b.String()
}
