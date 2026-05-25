### Senior-Engineer Report: landing-page
**Date:** 2026-05-25

#### Files Touched
| Path | Reason |
|------|--------|
| `web/index.html` | The entire deployable site: one self-contained HTML file with an inline `<style>` (terminal/brand-dark theme) and a single inline progressive-enhancement `<script>`. Hero, pipeline, greenfield, enforcement, demo, and footer sections; absolute OG/Twitter meta; inline SVG favicon. No build step, no external CSS/JS/CDN. |
| `web/assets/logo-banner.png` | Copied from repo-root `assets/` so `web/` is self-contained and maps 1:1 to the future gh-pages root. Referenced by the hero `<img>` with explicit `width="1280" height="640"` to avoid CLS. |
| `web/assets/social-preview.png` | Copied for the OG/Twitter social card. Referenced only by absolute URL in meta tags (the published gh-pages URL), so the file presence is for the gh-pages publish; not loaded by the page itself. |
| `web/assets/demo.gif` | Copied for the below-the-fold demo. Referenced with `loading="lazy"`, `decoding="async"`, explicit `width="1200" height="700"`, and meaningful `alt`, so it never blocks first paint and degrades to alt text on 404. |

Note: `assets/logo.png` (977 KB) was intentionally **not** copied — a small inline SVG (shield + eye) data-URI favicon is used instead, per the plan, to keep `web/` lean.

#### Architecture Compliance
- **Boundary checks passed:** This is a static marketing asset that lives **outside** the project's n-tier Go layers (`cmd/`, `internal/*`). It imports nothing, exports nothing, and participates in no Go package graph, so G2 (layer dependencies) is vacuously satisfied — there are no imports to cross a boundary.
- **G1 file size:** N/A by gate scope. The G1 file-size scanner (`internal/gates/file_size_scan.go`, `isSourceFile`) only walks code extensions (`.go .ts .tsx .js .jsx .py .rb .rs .java .kt .cs .cpp .c .h .swift .gd`). `.html`, `.css`, `.png`, and `.gif` are never scanned, so `web/index.html` (446 lines) cannot trip G1 regardless of length. No `file_size_exceptions` entry was added (none is needed), and **no JS lives in a scanned `.js`/`.ts` file** — all progressive-enhancement script is inline in `index.html`.
- **G7 outer-layer rule:** N/A. The page is pure presentation. The only script is progressive enhancement (clipboard copy + scroll-reveal); it carries no domain/business logic, persists nothing, and the page is fully legible and complete with JS disabled.

#### Type-Safety Notes
- N/A for static HTML — there is no typed source under the project's static-analysis tooling here.
- The document is valid HTML5: `<!doctype html>`, single `<h1>`, semantic `<header>/<nav>/<main>/<section>/<footer>`, ordered `<ol>` for the pipeline, and `alt`/`aria-label` on imagery and the terminal "image".
- No inline-event-handler sprawl: there are **no** `onclick`/`onload`-style attributes anywhere; the one listener is attached in the single inline `<script>` and guarded with feature detection (`navigator.clipboard`, `IntersectionObserver`, `matchMedia`), so unsupported environments no-op cleanly.

#### Trade-Offs
- **Inline SVG/CSS graphics over new image assets.** The pipeline, roadmap tree, and enforcement terminal are hand-built with CSS (flex nodes, chips, traffic-light dots) plus a tiny SVG favicon — no new artwork, no extra network requests, fully themeable, and crisp at any DPR. Rejected: shipping pre-rendered PNGs (adds weight, can't restyle, blurs on zoom).
- **`demo.gif` kept on the page (~901 KB)** rather than swapped for a static screenshot. It is below the fold with `loading="lazy"` + `decoding="async"` + reserved box (explicit dimensions), so the hero paints instantly and the GIF never blocks first paint; the motion is the strongest single proof of the tool, justifying the deferred weight.
- **`html.js`-gated reveal animations.** The initially-hidden `.reveal` opacity is scoped to `html.js` (added by the script only when motion is welcome). This guarantees the no-JS and reduced-motion paths render every section fully opaque — no content can be hidden behind an animation that never fires.
- **Words wrapped in `<span>` inside the value prop** to color the arrows/"enforced". The element's `textContent` still collapses to the exact canonical string `plan → code → tests → validate → docs — enforced` (verified), so DOM-based acceptance assertions pass.
- **Single sticky-nav + skip link** for keyboard/AT users; nav section links hidden on small screens to avoid crowding, brand+star always visible.

#### Handoff
- **Next role:** ux-ui-specialist
- **Outstanding TODOs:**
  - ux-ui-specialist to do the mobile-first/responsive + a11y review pass (320–360px no-overflow, AA contrast spot-check, reduced-motion, loading/empty/error states) and record `mobileFirst: true` with the eight required UX tags.
  - Out-of-band (maintainer): publish `web/` contents to an orphan `gh-pages` branch with `.nojekyll`, enable Pages, then `gh repo edit --homepage https://samuelnp.github.io/centinela/`.

---

#### Palette Change (user request, code-review pause, 2026-05-25)

User requested the accent be changed from the neon-green to the **blue used in the
logo**. Colors were sampled directly from `assets/logo-banner.png` (canvas pixel
sampling) — the shield is a cyan→azure gradient. Applied to `web/index.html`:

- `--accent` `#3fff9f` → `#40a0e0` (logo azure); `--accent-2` `#39d353` → `#30c0d0`
  (logo cyan); `--accent-ink` `#bdffdd` → `#bfe4ff`.
- All accent glows `rgba(63,255,159,…)` → `rgba(64,160,224,…)` and the secondary
  glow `rgba(57,211,83,…)` → `rgba(48,192,208,…)`.
- Dark-on-accent text `#04140b`/`#021008` → `#021420`/`#01101c` (blue-black) for hue
  harmony; inline SVG favicon stroke/fill re-colored to the azure.

Re-verified after the swap (Playwright): accent-as-text contrast **6.75:1** and
dark-text-on-button **6.52:1** (both pass WCAG AA on `--bg`); button gradient renders
azure→cyan matching the shield. The terminal traffic-light dot `#27c93f` was
**intentionally left green** — it is a skeuomorphic macOS-window control (red/amber/
green), not the brand highlight.
