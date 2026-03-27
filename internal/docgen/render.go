package docgen

import (
	"fmt"
	"html/template"
	"strings"
)

func RenderHTML(d *Data) string {
	b := &strings.Builder{}
	fmt.Fprint(b, "<!doctype html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>")
	fmt.Fprint(b, esc(d.Title))
	fmt.Fprint(b, "</title>", renderHeadAssets())
	fmt.Fprint(b, "</head><body><div class=\"page\">", renderNav(d), "<main class=\"content\">")
	fmt.Fprint(b, renderHero(d), renderOverview(d), renderFeatureGraphs(d))
	fmt.Fprint(b, renderRoadmapNarrative(d), renderSpecs(d), renderWorkflowState(d))
	fmt.Fprint(b, renderArtifacts(d), renderExamples(), renderSources(d))
	fmt.Fprint(b, "</main></div></body></html>")
	return b.String()
}

func esc(s string) string { return template.HTMLEscapeString(s) }

func listHTML(items []string) string {
	b := &strings.Builder{}
	b.WriteString("<ul>")
	for _, i := range items {
		b.WriteString("<li><code>" + esc(i) + "</code></li>")
	}
	b.WriteString("</ul>")
	return b.String()
}
