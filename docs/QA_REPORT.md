# SparkWA docs QA report

Date: 2026-11-06

## Change scope
A docs-only static information architecture for the existing Go repository, retaining its original workflows. The live GitHub Pages artifact is the `docs/` directory. No runtime OAuth, messaging, auth, billing, or backend behavior was introduced or changed.

## Implemented
- Source-backed product homepage and architecture explanation
- Structured docs overview and five rendered guide routes
- Preserved original complete reference in `reference/complete-guide.html`
- Search index over allowlisted public guide bodies and titles
- Ctrl/Cmd+K, '/' shortcut, keyboard search, focus trapping and Escape
- Dark/light preference toggle, responsive navigation, docs sidebar and heading TOC
- Code copy with clipboard success/error state
- No-JS HTML content, sitemap and canonical metadata
- Original SparkWA art and mark, responsive styles and print layout
- Section-18 effects audit in DESIGN_SYSTEM.md with per-effect status

## Automated test results
Results must be read from the GitHub Actions workflow(s) for the actual commit, not assumed here.

## Items requiring evidence before asserting success
- Static site structure, JavaScript syntax, internal routes, canonical links, and blocked secrets: **not yet executed**
- 320 / 360 / 390 / 768 / 1024 / 1280 / 1440 / 1920px screenshots and overflow tests: **not yet executed**
- Search, keyboard, menu, theme, TOC, copy states: **not yet executed**
- Reduced motion, blocked autoplay (no video), slow network and broken media fallbacks: **not yet executed**
- Lighthouse mobile/desktop performance, accessibility, SEO and best practices: **not measured**
- WCAG 2.2 AA: **not certified**
- Go build, Go tests and lints: **not rerun for this documentation revision**
- GitHub Pages deployment: **not yet observed**

## Operational cautions
Copying and running the documented installer with `sudo` is a privileged action; review the script before running. The docs discuss third-party WhatsApp and Gmail services without promising official endorsement or uptime. Live outbound URLs and vendor behavior may change.
