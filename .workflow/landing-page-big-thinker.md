### Big-Thinker Report: landing-page
**Date:** 2026-05-25

#### Problem
Centinela's GitHub `homepageUrl` currently points at `https://github.com/samuelnp/centinela#readme`, so every visitor — arriving from Hacker News, Reddit, the Claude/OpenCode Discords, or an "awesome" list — lands on a long README and has to *read* to grasp the value. Developers evaluating AI-coding tooling decide in seconds; prose is too slow. The repo has strong hygiene but zero stars, and the bottleneck is comprehension speed at first contact. We need a single visual surface that makes the core idea — *plan → code → tests → validate → docs, enforced as a mechanical constraint* — obvious through graphics rather than text, and that sells the greenfield story hardest: **describe a project → Centinela generates a roadmap (phases → features) → advance feature-by-feature through the five enforced steps**, plus the enforcement "aha" of a prewrite hook blocking an out-of-order write. The target user is a developer already on Claude Code or OpenCode deciding whether Centinela is worth `go install`-ing; the secondary user is anyone resharing the link, for whom the social-preview card must look credible.

#### Scope
- **In (v1):**
  - One self-contained static `index.html` at the repo root, no build step, no JS framework, no CDN runtime dependency.
  - Terminal / brand-dark visual direction (dark bg, monospace accents, neon-green glow) consistent with `assets/logo-banner.png`.
  - Above-the-fold hero: brand (logo + name), one-line value prop (`plan → code → tests → validate → docs — enforced`), the canonical `go install github.com/samuelnp/centinela@latest` command (copy-friendly), and primary CTAs (Get Started / Star on GitHub).
  - A 5-step **pipeline diagram** rendered in CSS/SVG, showing active vs. pending vs. done step states.
  - A **greenfield section** visually narrating describe-project → roadmap (phases → features) → feature-by-feature advancement.
  - An **enforcement "aha"** block: a styled depiction of the prewrite hook blocking a code write during the `plan` step.
  - Reuse of existing assets in `assets/` (logo-banner, demo.gif, social-preview, logo) — referenced, not regenerated.
  - Correct Open Graph + Twitter card meta with **absolute** URLs pointing at the published `assets/social-preview.png`.
  - Responsive/mobile-first layout, `prefers-reduced-motion` honored, WCAG AA contrast, real outbound links (repo, releases, README, HOWTO).
  - Publish to a dedicated `gh-pages` branch and repoint `homepageUrl` to `https://samuelnp.github.io/centinela/`.
- **Out (v1):**
  - Any build tooling, bundler, framework, or runtime JS dependency (animations are progressive enhancement only; the page must be fully legible with JS disabled).
  - New/regenerated artwork or a custom domain (use the `github.io` URL and existing assets).
  - Analytics, cookie banners, forms, newsletter signup, or any server/backend.
  - Auto-syncing copy from the README (kept minimal and linked instead — see maintenance-drift risk).
  - Multi-page docs site, blog, or interactive playground.

#### Dependencies & Assumptions
- **Existing assets (verified present in `assets/`):** `logo-banner.png` (57 KB), `logo.png` (954 KB), `social-preview.png` (259 KB), `demo.gif` (901 KB), `demo.tape`. These ship to `gh-pages` alongside `index.html`; the GIF is the only heavy item and must lazy-load.
- **Canonical copy is the README:** value prop is `Plan → code → tests → validate → docs — enforced.`; install command is `go install github.com/samuelnp/centinela@latest`; feature list and pipeline semantics already exist in the README's "How Centinela Works" mermaid and "Latest Features" list. Source landing copy from these to avoid drift.
- **G1 vs. HTML — RESOLVED, no carve-out needed.** The G1 file-size gate only walks files whose extension is in `isSourceFile` (`internal/gates/file_size_scan.go`: `.go .ts .tsx .js .jsx .py .rb .rs .java .kt .cs .cpp .c .h .swift .gd`). `.html`, `.css`, `.png`, and `.gif` are **never** scanned, so a multi-hundred-line `index.html` cannot trip G1 regardless of where it lives in the repo. `ignoreDirs` (`.git node_modules vendor dist .next target build`) is irrelevant here because the extension filter already excludes the file. **Recommendation:** keep `index.html` at the repo root and reuse the existing `assets/` directory — no new validator configuration, no `file_size_exceptions` entry, and no separate site directory required. Treat the page as a marketing asset, not application source; do not add inline `<script>` blocks in `.js`/`.ts` files (which *would* be scanned) — keep all progressive-enhancement script inline in the `.html` or omit it.
- **gh-pages mechanics:** publishing is a branch operation. Create an orphan `gh-pages` branch (`git switch --orphan gh-pages`) containing only `index.html` + `assets/` (+ a `.nojekyll` file so GitHub Pages serves the page verbatim without Jekyll processing), push it, enable Pages → branch `gh-pages` → root in repo settings, confirm `https://samuelnp.github.io/centinela/` resolves, then repoint `homepageUrl` via `gh repo edit samuelnp/centinela --homepage https://samuelnp.github.io/centinela/`. Assumes the maintainer has push + repo-admin rights (Pages config + homepage edit are out-of-band, maintainer-run steps).
- **`main` stays clean:** the source `index.html` lives on `main` (as the feature deliverable); `gh-pages` is the publish target. The plan documents the publish procedure so `main` history is unaffected.
- **No i18n:** project is English-only (`gates.i18n = false`), so G11 does not apply to the page.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Page weight from `demo.gif` (~901 KB) hurts first paint / mobile load | Medium | Medium | `loading="lazy"` + `decoding="async"`, place GIF below the fold, give it explicit width/height + a lightweight poster/placeholder so layout does not shift; hero must paint without it. |
| G1 file-size rule misapplied to the HTML file blocking validate | Low | Low | Confirmed `isSourceFile` excludes `.html`/`.css`/images — gate never scans them; keep all script inline in `.html` (never in a scanned `.js`/`.ts`); document this in the plan so no one adds a needless `file_size_exceptions` entry. |
| Social-preview card breaks on Slack/Discord/X (relative OG URLs) | Medium | Medium | Use **absolute** `https://samuelnp.github.io/centinela/assets/social-preview.png` in `og:image` and `twitter:image`; set `og:url`, `og:type`, `twitter:card=summary_large_image`; validate with a card debugger after publish. |
| Maintenance drift: hardcoded version/feature copy diverges from README | Low | Medium | Keep copy minimal, prefer evergreen value prop over version numbers, use `@latest` install string, link "full feature list" to README as canon rather than duplicating it. |
| Narrow viewport (≤360px) overflow on pipeline/roadmap graphics | Medium | Medium | Mobile-first CSS; pipeline/roadmap reflow from horizontal to vertical stacks below a breakpoint; test at 320–360px; no fixed-width SVG that forces horizontal scroll. |
| Animation reliance breaks no-JS / reduced-motion users | Low | Low | All content legible without JS; animations are CSS progressive enhancement gated behind `prefers-reduced-motion: no-preference`. |
| Contrast fails WCAG AA on neon-green-on-dark | Low | Medium | Reserve neon green for accents/large text/glow; use a high-contrast off-white for body copy; verify AA ratios. |
| Dead/incorrect outbound links | Low | Low | Link only to known-real targets (repo, `/releases`, `/blob/main/README.md`, `/blob/main/HOWTO.md`, `/stargazers`); spot-check after publish. |
| gh-pages publish disturbs `main` or loses assets | Low | Low | Use an orphan branch + `.nojekyll`; document exact publish commands; keep source on `main` so a rebuild is reproducible. |

#### Rollout
- **Step 1 — Shell slice (smallest correct slice):** `index.html` skeleton with the terminal/brand-dark CSS system (inline `<style>`), the hero (logo-banner, name, value prop, copy-friendly `go install` command, Get Started + Star CTAs), the OG/Twitter meta block with absolute URLs, and the responsive baseline + reduced-motion guard. This alone is a shippable landing page.
- **Step 2 — Pipeline graphic:** add the 5-step `plan → code → tests → validate → docs` diagram (CSS/SVG) with active/pending/done states, reflowing vertically on narrow viewports.
- **Step 3 — Greenfield + enforcement graphics:** add the describe→roadmap(phases→features)→feature-by-feature section and the prewrite-hook "blocked write" enforcement panel; lazy-load `demo.gif` here below the fold.
- **Step 4 — Publish + repoint:** create orphan `gh-pages` with `index.html` + `assets/` + `.nojekyll`, push, enable GitHub Pages, verify `https://samuelnp.github.io/centinela/`, validate the social card, then `gh repo edit --homepage` to repoint `homepageUrl`. (Out-of-band, maintainer-run.)

#### Handoff
- **Next role:** feature-specialist
- **Outstanding questions:**
  - Should the pipeline/greenfield visuals be hand-authored inline SVG (sharper, themeable, no extra asset) or a new static image? Recommendation: inline SVG/CSS to stay self-contained and avoid new artwork (which is out of scope).
  - Confirm the maintainer will run the two out-of-band steps (Pages enablement + `gh repo edit --homepage`) — these need repo-admin rights and are not automatable inside the feature branch.
  - Decide whether `demo.gif` appears on the landing page at all (weight cost) or is replaced by a static screenshot + "see the demo on the README" link to cut payload further.
