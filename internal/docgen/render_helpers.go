package docgen

import "strings"

func firstLines(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	parts := strings.Split(s, "\n")
	if len(parts) <= n {
		return s
	}
	return strings.Join(parts[:n], "\n")
}

func roadmapList(nodes []RoadmapNode) string {
	if len(nodes) == 0 {
		return "<p>No roadmap nodes available.</p>"
	}
	b := &strings.Builder{}
	b.WriteString("<ul class=\"list\">")
	for _, n := range nodes {
		dep := "none"
		if len(n.DependsOn) > 0 {
			dep = strings.Join(n.DependsOn, ", ")
		}
		b.WriteString("<li><strong>" + esc(n.Name) + "</strong> <span class=\"k\">depends on: " + esc(dep) + "</span></li>")
	}
	b.WriteString("</ul>")
	return b.String()
}

func renderExamples() string {
	return `<section id="examples"><h2>Documentation Examples</h2><p>Use this sequence for hybrid generation with LLM synthesis first.</p><div class="sample"><div class="sample-title">1) Validate artifacts</div><pre>centinela docs validate</pre></div><div class="sample"><div class="sample-title">2) Ask LLM to compose docs narrative + visual structure</div><pre>Use docs/architecture/documentation-generator-prompt.md to generate project-facing docs narrative and visuals.</pre></div><div class="sample"><div class="sample-title">3) Deterministic fallback renderer</div><pre>centinela docs generate --out docs/project-docs/index.html</pre></div></section>`
}
