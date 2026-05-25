---
feature: landing-page
summary: A fast, visual landing page that explains Centinela at a glance and links to install.
audience: end-user
status: done
---

## What it does
Centinela now has a polished public landing page at https://samuelnp.github.io/centinela/ with a terminal/brand-dark aesthetic that matches the logo. It shows the one-line value prop (plan → code → tests → validate → docs — enforced), the install command, a visual pipeline of the 5-step workflow, the greenfield story (describe project → Centinela builds a roadmap → advance feature by feature), and a panel showing the enforcement mechanism in action. It is the repo's official homepage.

## When you'd use it
When someone discovers Centinela from a link (Hacker News, Reddit, an "awesome" list, social media, or word-of-mouth) and wants to understand what it does in seconds before deciding to install or star it. You'd share this link with colleagues evaluating AI-coding frameworks.

## How it behaves
- Shows a full, interactive hero section with the Centinela logo banner, project name, and the core value prop immediately visible when the page loads.
- Displays the exact `go install github.com/samuelnp/centinela@latest` install command in a code block, ready to copy.
- Includes primary call-to-action buttons ("Get Started" and "Star on GitHub") above the fold for immediate action.
- Lays out the five workflow steps (plan, code, tests, validate, docs) as a visual pipeline diagram with clear styling showing which steps are active, pending, or complete.
- Illustrates the greenfield workflow with phases and features, explaining how describing a project leads to a generated roadmap that you advance through one enforced feature at a time.
- Depicts the enforcement mechanism with a "blocked write" panel showing how the pre-write hook prevents stepping out of order—proving enforcement is real, not just a suggestion.
- Includes social-preview metadata (Open Graph and Twitter Card tags with absolute URLs) so the page renders beautifully when shared on Slack, Discord, Twitter, or other platforms.
- Self-hosts all assets (logo, social preview image, demo GIF) without external CDN or JavaScript framework dependencies—truly zero build step and zero runtime network calls.
- Lazy-loads the demo GIF with explicit dimensions and a placeholder so the page paints fast and never blocks first paint, even on slow connections.
- Contains real, verified outbound links in the footer (repository, releases, README, HOWTO, and license).
- Remains fully readable and legible when JavaScript is disabled, with all content and layout intact in no-JS mode.
- Respects the `prefers-reduced-motion` OS setting, disabling all CSS animations when the user requests reduced motion while keeping all content visible.
- Reflows responsively on mobile devices (320–360px width) with the pipeline diagram stacking vertically and no horizontal scrolling or clipped content.
- Gracefully degrades if the demo GIF fails to load, showing the alt text and maintaining layout without shifting or throwing JavaScript errors.

## Examples
Visit https://samuelnp.github.io/centinela/ to see the landing page in action. The page is instantly shareable—paste the link in Slack or Twitter and the preview card appears automatically with the correct image and description.
