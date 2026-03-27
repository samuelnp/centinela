package docgen

import (
	"fmt"
	"html/template"
	"strings"
)

func RenderHTML(d *Data) string {
	b := &strings.Builder{}
	fmt.Fprint(b, "<!doctype html><html><head><meta charset=\"utf-8\"><title>")
	fmt.Fprint(b, esc(d.Title))
	fmt.Fprint(b, "</title><script type=\"module\">import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs';mermaid.initialize({startOnLoad:true});</script></head><body>")
	fmt.Fprint(b, "<h1>", esc(d.Title), "</h1><h2>How Project Works</h2><pre>", esc(d.Project), "</pre>")
	fmt.Fprint(b, "<h2>Roadmap Narrative</h2><pre>", esc(d.RoadmapText), "</pre>")
	fmt.Fprint(b, "<h2>Mermaid: Feature Dependencies</h2><pre class=\"mermaid\">", mermaidRoadmap(d.RoadmapNodes), "</pre>")
	fmt.Fprint(b, "<h2>Mermaid: Evidence Handoffs</h2><pre class=\"mermaid\">", mermaidEvidence(d.Evidence), "</pre>")
	fmt.Fprint(b, "<h2>Comparison Matrix</h2>", statesTable(d.States))
	fmt.Fprint(b, "<h2>Specs</h2><p>Files: ", len(d.Specs), " · Scenarios: ", d.Scenarios, "</p>", listHTML(d.Specs))
	fmt.Fprint(b, "<h2>Feature Docs</h2>", listHTML(d.FeatureDocs), "<h2>Plan Docs</h2>", listHTML(d.PlanDocs))
	fmt.Fprint(b, "<h2>Evidence to Code References</h2>", evidenceTable(d.Evidence), "</body></html>")
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
