# Feature: Centinela Landing Page

> surface: user-facing
> deliverable: static marketing site (self-contained HTML + assets), hosted on the `gh-pages` branch
> homepageUrl target: https://samuelnp.github.io/centinela/

## Problem

Centinela's GitHub `homepageUrl` currently points at `#readme`. A visitor arriving
from Hacker News, Reddit, the Claude/OpenCode Discords, or an "awesome" list lands
on a long README and has to *read* to understand the value. There is no fast,
visual "what is this and why should I care" surface.

The repo has zero stars despite strong hygiene. The bottleneck is **comprehension
speed at first contact**: a developer evaluating AI-coding tooling decides in
seconds. We need a single page that makes Centinela's core idea —
*plan → code → tests → validate → docs, enforced as a mechanical constraint* —
obvious through graphics, not prose, and that specifically sells the greenfield
story: **describe a project → Centinela builds a roadmap → you advance through it
one enforced feature at a time.**

**Who is the user?** Developers already using Claude Code or OpenCode who are
deciding whether Centinela is worth `go install`-ing. Secondary: people sharing
the link (the social-preview card must look good).

## User Stories

- As a developer evaluating Centinela, I want to understand the core idea in under
  10 seconds from visuals alone, so that I decide to keep reading instead of bouncing.
- As a greenfield builder, I want to see how roadmap creation + step-by-step
  advancement works, so that I trust Centinela to scaffold a new project end-to-end.
- As a skeptic, I want to see the *enforcement* "aha" (a hook blocking an
  out-of-order write), so that I believe it is a real constraint, not a suggestion.
- As a visitor ready to act, I want a one-line install command and a GitHub/Star
  CTA above the fold, so that I can adopt or star immediately.
- As someone on mobile or sharing the link, I want the page to render well on a
  phone and produce a rich social-preview card, so that it is credible when shared.

## Acceptance Criteria

1. A single self-contained landing page (`index.html` + referenced assets) renders
   with **no build step and no external JS framework or CDN runtime dependency**.
2. Above the fold shows: the brand (logo + name), the one-line value prop
   (`plan → code → tests → validate → docs — enforced`), the `go install` command,
   and primary CTAs (Get Started / Star on GitHub).
3. A **visual pipeline diagram** of the 5 steps is present, styled with the
   terminal/brand-dark aesthetic, showing the active-step / pending-step states.
4. A **greenfield section** visually explains: describe project → roadmap generated
   (phases → features) → advance feature-by-feature through the 5 steps.
5. An **enforcement "aha"** is shown: a depiction of the prewrite hook blocking a
   code write during the `plan` step.
6. The terminal/brand-dark visual direction is applied (dark background, monospace
   accents, neon-green glow), consistent with `assets/logo-banner.png`.
7. The page is responsive (mobile-first) and includes correct social-preview
   (Open Graph + Twitter card) meta tags pointing at `assets/social-preview.png`.
8. All outbound links resolve to real targets (repo, releases, README, HOWTO,
   install instructions). No dead links.
9. The page is hostable on the `gh-pages` branch at the documented URL and is what
   `homepageUrl` will point to.

## Edge Cases

- **No JavaScript** / JS disabled: content and layout must remain fully legible
  (animations are progressive enhancement only).
- **Reduced motion**: honor `prefers-reduced-motion`; disable non-essential animation.
- **Narrow viewport (≤360px)**: pipeline diagram and roadmap graphic must reflow,
  not overflow horizontally.
- **Dark/light OS preference**: page is intentionally dark; ensure contrast meets
  WCAG AA regardless of OS theme.
- **Missing/slow asset**: `demo.gif` is ~900KB — must lazy-load and not block first
  paint; provide a poster/placeholder.
- **Broken share preview**: social-preview meta must use absolute URLs so cards
  render on Slack/Discord/X.
- **Stale version/install string**: install command must match the README's
  canonical `go install github.com/samuelnp/centinela@latest`.

## Data Model

No persistent data. Static content only. Conceptual entities depicted (not stored):
- **Step** — one of `plan|code|tests|validate|docs`, each with a state (`active|pending|done`).
- **Roadmap** — ordered list of **Phases**, each containing **Features**.
- **Hook event** — a blocked write with a reason message (for the enforcement visual).

## Integration Points

- **GitHub Pages** (`gh-pages` branch) — hosting surface.
- **GitHub repo** — links to releases, README, HOWTO, stargazers.
- **Existing assets** — `assets/logo-banner.png`, `assets/demo.gif`,
  `assets/social-preview.png`, `assets/logo.png` (reused, not regenerated).
- **`homepageUrl`** — repo metadata field repointed after the page ships
  (out of band, via `gh repo edit`).

## Risks

- **Performance / page weight** (Medium): large GIF could hurt first paint. Mitigate
  with lazy-loading and deferred decode.
- **G1 file-size rule vs. HTML** (Medium): a single HTML file far exceeds the 100-line
  source rule. Mitigation: this is a marketing asset, not application source under an
  enforced layer; confirm `validate` scope excludes the site directory (or treat it as
  a non-`.go` asset the gate does not scan). To be settled in the plan.
- **Maintenance drift** (Low): hardcoded copy (version, feature list) can fall out of
  sync with the README. Mitigation: keep copy minimal and link to the README as canon.
- **Branch hygiene** (Low): publishing on `gh-pages` requires keeping `main` clean and
  a clear publish procedure.

## Decomposition

Small enough to ship as one feature. If it grows, candidate slices:
- `landing-page-shell` — HTML skeleton, hero, install, CTAs, meta/OG tags.
- `landing-page-pipeline-graphic` — the 5-step enforced pipeline visual.
- `landing-page-greenfield-graphic` — roadmap creation + advancement visual.
- `landing-page-publish` — `gh-pages` branch + Pages config + `homepageUrl` repoint.
