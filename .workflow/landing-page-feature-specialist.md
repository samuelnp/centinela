### Feature-Specialist Report: landing-page
**Date:** 2026-05-25

#### Behavior Summary

A visitor who navigates to `https://samuelnp.github.io/centinela/` receives a single self-contained `index.html` with no external runtime JS or CSS dependencies. The page opens on a terminal/brand-dark hero that shows the Centinela logo-banner, the name, the one-line value prop (`plan → code → tests → validate → docs — enforced`), a copy-friendly `go install github.com/samuelnp/centinela@latest` command block, and two primary CTAs (Get Started, Star on GitHub) — all above the fold and fully legible with JavaScript disabled. Scrolling reveals: (1) a visual pipeline diagram of the five enforced steps with `active / pending / done` state styling that reflows to a vertical stack on narrow viewports; (2) a greenfield section narrating describe-project → roadmap-of-phases-and-features → feature-by-feature advancement through the five steps; (3) an enforcement "aha" panel depicting the prewrite hook blocking a source-code write during the `plan` step with a real reason message; (4) a lazy-loaded `demo.gif` with an explicit width/height and lightweight placeholder so it never blocks first paint; and (5) a footer with verified outbound links (repo, releases, README, HOWTO, license). The `<head>` carries complete Open Graph and Twitter card meta tags using **absolute** URLs pointing at the published `assets/social-preview.png`, so sharing on Slack/Discord/X produces a rich preview card. All CSS animations are guarded behind `prefers-reduced-motion: no-preference`; the base experience is static. The page is mobile-first and meets WCAG AA contrast on its intentional dark theme.

#### Gherkin Scenarios

See `specs/landing-page.feature` for the full executable spec.

1. **Happy path — full interactive render**: visitor opens the page with JS enabled; all five required above-the-fold elements are present; pipeline, greenfield, and enforcement sections render; GIF loads lazily; footer links are real.
2. **No-JS / progressive-enhancement degraded path**: visitor has JS disabled; page layout and all content remain fully legible; no JS-only content is lost.
3. **Reduced-motion preference**: visitor's OS has `prefers-reduced-motion: reduce`; no CSS animations play; all static content is still present.
4. **Narrow viewport (≤360px)**: at 360px width the pipeline diagram and roadmap graphic reflow to vertical stacks without horizontal scroll.
5. **Install command matches README canonical string**: the `go install` command in the page exactly matches `go install github.com/samuelnp/centinela@latest`.
6. **Absolute OG and Twitter meta tags**: `og:image` and `twitter:image` values are absolute URLs starting with `https://samuelnp.github.io/centinela/assets/social-preview.png`.
7. **No external runtime CDN dependency**: the rendered HTML contains no `<script src>` or `<link rel="stylesheet" href>` pointing at an external host.
8. **demo.gif has lazy-load attributes**: the `<img>` element for `assets/demo.gif` carries `loading="lazy"` and `decoding="async"` and explicit `width`/`height` attributes.
9. **Missing/slow demo.gif — placeholder present**: the GIF element has a visible poster or placeholder background so layout does not shift when the asset is slow or missing.
10. **All required above-the-fold elements present**: the hero contains the logo-banner image, the text "Centinela", the value prop string, the install command block, and both CTA links.
11. **Footer links are non-empty and real-target hrefs**: every `<a>` in the footer carries an `href` attribute pointing at a known GitHub URL (repo, releases, README, HOWTO, license), not `#` or empty.
12. **Enforcement "aha" panel present**: a block depicting the prewrite hook blocking a write is present on the page.
13. **Greenfield roadmap section present**: a section visually depicting describe-project → roadmap-phases-features → feature-by-feature advancement is present.

#### UX States

| State   | Trigger | Surface |
|---------|---------|---------|
| loading | Browser first-paint before `demo.gif` finishes fetching | `<img>` element for `demo.gif` shows a lightweight placeholder/poster (background color or low-res stand-in); layout does not shift because explicit `width`/`height` are set; hero paints immediately without waiting for the GIF. |
| empty   | JS disabled (no progressive-enhancement JS runs) | All page content — hero, pipeline, greenfield, enforcement, footer — renders from static HTML/CSS only; no section collapses or disappears; the copy-button (if any) silently absent or non-functional without degrading layout. |
| error   | `demo.gif` fails to load (404, timeout, or network loss) | `<img>` alt text is shown; the explicit `width`/`height` preserves layout; the rest of the page is unaffected; no JS error modal. |
| success | Page fully loaded with all assets and JS active | Hero, pipeline diagram (with step state styles), greenfield section, enforcement panel, and GIF all rendered; copy button (if present) functional; all outbound links correct. |

#### Out-of-Scope

- Any build tooling, bundler, or JS framework (Vite, webpack, React, Vue, etc.).
- External CSS/JS CDN runtime dependencies (no Bootstrap CDN, no Google Fonts, no analytics scripts).
- New or regenerated artwork — existing `assets/` files are reused as-is.
- A custom domain — the page is served at the `github.io` URL only.
- Auto-syncing copy from the README (copy is minimal and links to README as canon).
- Analytics, cookie banners, newsletter signup, or any server/backend.
- Multi-page docs site, blog, or interactive playground.
- Internationalisation (project is English-only; `gates.i18n = false`).
- A separate `.js` or `.css` file on `main` whose extension would be scanned by G1 — all script stays inline in `index.html`.
- Enabling GitHub Pages or repointing `homepageUrl` (these are out-of-band maintainer-run steps documented in the plan, not automated by this feature branch).

#### Handoff

- **Next role:** senior-engineer
- **Open clarifications:**
  - Confirm whether `demo.gif` appears on the page or is replaced by a static screenshot + "see the demo" README link to cut ~900 KB payload; the plan notes this as a decision for the feature-specialist — recommendation is to include it with `loading="lazy"` + placeholder, but the senior-engineer should confirm final weight budget.
  - Confirm that the maintainer will run the two out-of-band publish steps (Pages enablement + `gh repo edit --homepage`) once the branch ships.
