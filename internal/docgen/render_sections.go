package docgen

import "fmt"

func renderFeatureGraphs(d *Data) string {
	return `<section id="feature-graphs"><h2>Feature Topology</h2><p>These Mermaid diagrams explain project feature scope and specification surface.</p><h3>Roadmap Dependencies</h3><div class="mermaid-wrap"><pre class="mermaid">` + mermaidRoadmap(d.RoadmapNodes) + `</pre></div><h3>Spec Coverage Map</h3><div class="mermaid-wrap"><pre class="mermaid">` + mermaidSpecs(d.Specs) + `</pre></div></section>`
}

func renderRoadmapNarrative(d *Data) string {
	return `<section id="roadmap"><h2>Roadmap Narrative</h2><p>Current roadmap intent, plus detected feature dependency context.</p><div class="code-note"><strong>Roadmap excerpt</strong><pre>` + esc(firstLines(d.RoadmapText, 20)) + `</pre></div>` + roadmapList(d.RoadmapNodes) + `</section>`
}

func renderSpecs(d *Data) string {
	return fmt.Sprintf(`<section id="specs"><h2>Specification Inventory</h2><p>Files: %d · Scenarios: %d</p>%s</section>`, len(d.Specs), d.Scenarios, listHTML(d.Specs))
}

func renderWorkflowState(d *Data) string {
	return `<section id="workflow-state"><h2>Feature State Matrix</h2><p>Operational status table for tracked features.</p>` + statesTable(d.States) + `</section>`
}

func renderArtifacts(d *Data) string {
	return `<section id="artifacts"><h2>Artifacts</h2><h3>Feature briefs</h3>` + listHTML(d.FeatureDocs) + `<h3>Implementation plans</h3>` + listHTML(d.PlanDocs) + `<h3>Evidence to code references</h3>` + evidenceTable(d.Evidence) + `</section>`
}

func renderSources(d *Data) string {
	return `<section id="sources"><h2>Source Context</h2><details><summary>PROJECT.md excerpt</summary><pre>` + esc(firstLines(d.Project, 28)) + `</pre></details><details><summary>ROADMAP.md excerpt</summary><pre>` + esc(firstLines(d.RoadmapText, 28)) + `</pre></details></section>`
}
