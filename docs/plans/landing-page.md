# Plan: Centinela Landing Page

A single self-contained static landing page (`index.html` + reused `assets/`),
terminal / brand-dark aesthetic, no build step and no runtime JS dependency,
published on a dedicated `gh-pages` branch and set as the repo `homepageUrl`
target at `https://samuelnp.github.io/centinela/`.

## File & directory layout

- `index.html` — repo root. The entire page: one HTML file with an inline
  `<style>` block (terminal/brand-dark theme) and, optionally, one small inline
  `<script>` for progressive-enhancement only. No external CSS/JS, no CDN.
- `assets/` — reuse existing files, do not regenerate:
  `logo-banner.png`, `logo.png`, `social-preview.png`, `demo.gif`.
- `.nojekyll` — added on the `gh-pages` branch only, so GitHub Pages serves
  the page verbatim (no Jekyll processing).
- G1 note: the G1 file-size gate (`internal/gates/file_size_scan.go`,
  `isSourceFile`) only scans code extensions (`.go .ts .tsx .js .jsx .py .rb
  .rs .java .kt .cs .cpp .c .h .swift .gd`). `.html`/`.css`/images are never
  scanned, so a long `index.html` cannot trip G1. Do **not** add a
  `file_size_exceptions` entry and do **not** put page script in a scanned
  `.js`/`.ts` file — keep all script inline in `index.html`.

## `index.html` section order (top to bottom)

1. `<head>` meta block — charset, viewport (`width=device-width,
   initial-scale=1`), `<title>`, `<meta name="description">`, theme-color, and
   the social-preview tags using **absolute** URLs:
   - `og:title`, `og:description`, `og:type=website`,
     `og:url=https://samuelnp.github.io/centinela/`,
     `og:image=https://samuelnp.github.io/centinela/assets/social-preview.png`
   - `twitter:card=summary_large_image`, `twitter:title`,
     `twitter:description`, `twitter:image` (same absolute URL).
2. Inline `<style>` — the brand-dark theme: dark background, off-white
   high-contrast body text, neon-green accent reserved for highlights/glow,
   monospace accents for code/labels. Define a `prefers-reduced-motion:
   no-preference` block that owns all animation; base styles are static.
   Mobile-first: base layout is single-column; add min-width media queries for
   wider layouts.
3. **Hero (above the fold)** — `assets/logo-banner.png` (with width/height to
   avoid CLS), the name "Centinela", the one-line value prop
   `plan → code → tests → validate → docs — enforced`, a copy-friendly
   `go install github.com/samuelnp/centinela@latest` command block, and two
   primary CTAs: "Get Started" (→ repo README / install) and
   "Star on GitHub" (→ `https://github.com/samuelnp/centinela`).
4. **Pipeline diagram** — the 5 steps `plan → code → tests → validate → docs`
   as CSS/inline-SVG nodes with explicit `active | pending | done` visual
   states. Horizontal on desktop; reflows to a vertical stack at ≤360px (no
   horizontal overflow).
5. **Greenfield section** — visual narrative: *describe a project* → *Centinela
   generates a roadmap* (Phases → Features) → *advance feature-by-feature*
   through the 5 steps. Show the phase→feature hierarchy and the per-feature
   loop. Keep copy minimal; this is the hardest-selling section.
6. **Enforcement "aha"** — a styled terminal panel depicting the prewrite hook
   blocking a source-code write while the workflow is in the `plan` step (a
   "blocked write + reason message" mock), proving enforcement is a real
   constraint, not a suggestion.
7. **Demo (below the fold)** — `assets/demo.gif` with `loading="lazy"`,
   `decoding="async"`, explicit width/height, and a lightweight
   poster/placeholder so it never blocks first paint. (Or substitute a static
   screenshot + "see the demo" link to the README to cut payload — confirm with
   feature-specialist.)
8. **Footer** — real outbound links only: repo, Releases
   (`/releases/latest`), README (`/blob/main/README.md`), HOWTO
   (`/blob/main/HOWTO.md`), license. No dead links.

## Implementation steps (ordered)

1. Scaffold `index.html` head + meta (incl. absolute OG/Twitter tags) and the
   inline brand-dark `<style>` system (colors, type scale, monospace accents,
   reduced-motion guard, mobile-first base + breakpoints).
2. Build the **hero**: logo-banner, name, value prop, `go install` command
   block, Get Started + Star CTAs. Verify it paints fully with JS disabled.
3. Build the **pipeline diagram** (CSS/inline-SVG, active/pending/done states,
   reflow at ≤360px).
4. Build the **greenfield section** (describe → roadmap phases→features →
   feature-by-feature advancement) and the **enforcement "aha"** blocked-write
   panel.
5. Add the **demo** block (lazy-loaded GIF or static screenshot) and the
   **footer** with verified outbound links.
6. Accessibility/perf pass: WCAG AA contrast on dark theme, `prefers-reduced-
   motion` honored, viewport 320–360px has no horizontal overflow, hero paints
   before the GIF, no external network calls.

## Publish to `gh-pages` (out-of-band, maintainer-run)

1. Commit `index.html` on `main` (the source of record).
2. Create an orphan publish branch:
   `git switch --orphan gh-pages`
3. Keep only `index.html` + `assets/` on it, add an empty `.nojekyll`, commit,
   and `git push -u origin gh-pages`.
4. In GitHub repo settings → Pages, set source to branch `gh-pages`, folder
   `/ (root)`. Wait for the build, then confirm
   `https://samuelnp.github.io/centinela/` resolves and renders.
5. Validate the social card with a card debugger (Slack/Discord/X) to confirm
   the absolute `og:image` loads.
6. Return to `main`: `git switch main`.

## Repoint `homepageUrl` (out-of-band, maintainer-run)

After the page is live and verified:

```bash
gh repo edit samuelnp/centinela --homepage https://samuelnp.github.io/centinela/
```

(Current value is `https://github.com/samuelnp/centinela#readme`.)

## Acceptance traceability

Covers the feature brief's criteria 1–9: self-contained no-build page (1),
above-the-fold hero with value prop + install + CTAs (2), pipeline diagram with
step states (3), greenfield roadmap section (4), enforcement "aha" (5),
brand-dark aesthetic (6), responsive + OG/Twitter meta with absolute URLs (7),
real outbound links (8), and `gh-pages` hostability + `homepageUrl` repoint (9).
