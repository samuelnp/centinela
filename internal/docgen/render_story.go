package docgen

func renderLatestFeatures() string {
	return `<section id="latest-features"><h2>Latest Features</h2><p>Centinela's current user-facing capabilities focus on keeping setup, execution, validation, and documentation aligned across agent sessions.</p><div class="cards"><div class="card"><h3>Claude + OpenCode</h3><p>Shared prewrite enforcement, postwrite updates, setup prompts, and workflow context across both integrations.</p></div><div class="card"><h3>Roadmap-First Bootstrap</h3><p><code>centinela init</code>, roadmap generation, senior PM analysis, roadmap quality evaluation, and <code>centinela roadmap validate</code>.</p></div><div class="card"><h3>Five-Step Workflow</h3><p>Enforced <code>plan -&gt; code -&gt; tests -&gt; validate -&gt; docs</code> flow with configurable confirmation mode.</p></div><div class="card"><h3>Prompt-Driven UX</h3><p>Auto-start for new feature intent, strict orchestration directives, and a compact status line for active workflows.</p></div><div class="card"><h3>Managed Migration</h3><p><code>centinela migrate</code>, <code>centinela migrate docs</code>, and <code>centinela migrate setup --agent ...</code> keep managed assets current.</p></div><div class="card"><h3>Docs Output</h3><p><code>centinela docs validate</code> and <code>centinela docs generate --out docs/project-docs/index.html</code> publish the HTML project presentation.</p></div></div></section>`
}

func renderGettingStarted() string {
	return `<section id="getting-started"><h2>Getting Started</h2><p>Use this path to learn the Centinela workflow end to end.</p><div class="sample"><div class="sample-title">1) Bootstrap the project</div><pre>centinela init
centinela init --agent both</pre></div><div class="sample"><div class="sample-title">2) Complete setup and roadmap prerequisites</div><pre>PROJECT.md
ROADMAP.md
.workflow/roadmap-analysis.md + .json
.workflow/roadmap-quality.md + .json
centinela roadmap
centinela roadmap validate</pre></div><div class="sample"><div class="sample-title">3) Start and move one feature through the workflow</div><pre>centinela start my-feature
centinela complete my-feature   # plan -&gt; code
centinela complete my-feature   # code -&gt; tests
centinela complete my-feature   # tests -&gt; validate</pre></div><div class="sample"><div class="sample-title">4) Validate and publish docs</div><pre>centinela validate
centinela complete my-feature   # validate -&gt; docs
centinela docs validate
centinela docs generate --out docs/project-docs/index.html</pre></div></section>`
}
