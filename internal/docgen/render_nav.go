package docgen

import "fmt"

func renderNav(d *Data) string {
	return fmt.Sprintf(`<aside class="sidebar"><div class="brand">%s</div><div class="caption">Hybrid docs report</div><nav class="toc"><a href="#overview">Overview</a><a href="#latest-features">Latest Features</a><a href="#getting-started">Getting Started</a><a href="kb/index.html">Knowledge Base</a><a href="#feature-graphs">Feature Graphs</a><a href="#roadmap">Roadmap</a><a href="#specs">Specs</a><a href="#workflow-state">Feature States</a><a href="#artifacts">Artifacts</a><a href="#examples">Examples</a><a href="#sources">Source Context</a></nav><p class="caption">%d specs · %d scenarios · %d KB guides</p></aside>`, esc(d.Title), len(d.Specs), d.Scenarios, len(d.KB))
}

func renderHero(d *Data) string {
	return fmt.Sprintf(`<header class="hero"><h1>%s</h1><p class="meta">Generated from Centinela artifacts with a current workflow story, command surface, and traceability view for the project.</p><span class="pill">Latest features</span><span class="pill">Workflow onboarding</span><span class="pill">Mermaid topology</span></header>`, esc(d.Title))
}

func renderOverview(d *Data) string {
	return fmt.Sprintf(`<section id="overview"><h2>Project Overview</h2><p>Documentation quality is improved with navigable sections, digestible summaries, and richer visual presentation.</p><div class="cards"><div class="card"><div class="k">Feature Docs</div><div class="v">%d</div></div><div class="card"><div class="k">Plan Docs</div><div class="v">%d</div></div><div class="card"><div class="k">Specs</div><div class="v">%d</div></div><div class="card"><div class="k">Scenarios</div><div class="v">%d</div></div><div class="card"><div class="k">Tracked Features</div><div class="v">%d</div></div></div></section>`, len(d.FeatureDocs), len(d.PlanDocs), len(d.Specs), d.Scenarios, len(d.RoadmapNodes))
}
