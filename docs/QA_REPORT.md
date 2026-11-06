# SparkWA documentation site — QA report

**Date:** 2026-11-06  
**Site:** https://krishujeniya.github.io/sparkwa/  
**Scope:** Single-product public documentation; no backend/runtime changes.

## Implementation shipped
- Light-first, original SparkWA-branded homepage and product architecture explanation.
- Five allowlisted source-derived guide pages (VPS install, Gemini Spark, Gmail OAuth, security, troubleshooting).
- Documentation index with sidebar, breadcrumb, per-article table of contents, and previous/next links.
- Ctrl/Cmd+K and '/' searchable static public-content index, keyboard navigation, focus containment and Escape.
- Code-copy and SVG-logo copy with live status feedback; dark/light reading theme; responsive mobile navigation.
- Original full reference preserved at `/reference/complete-guide.html`.
- Semantic HTML, fixed-aspect media, lazy offscreen dependencies, reduced-motion CSS, print styles.
- Static GitHub Pages deployment, no added frontend runtime packages, real repo assets and GitHub links.

## Automated results: PASS

**GitHub Pages:** [deployment passed](https://github.com/krishujeniya/sparkwa/actions/runs/42158554305) for the final site content. Later commits adjust only QA scripts/workflows and this report.  
**Repository Go CI:** [passed](https://github.com/krishujeniya/sparkwa/actions/runs/42158936839) following the redesign.  
**Static documentation verification:** [passed](https://github.com/krishujeniya/sparkwa/actions/runs/42159046566): HTML routes, JS syntax, canonical URLs, local links, static assets, sitemap, search index, and 93-row effects audit.  
**Chromium browser verification:** [passed](https://github.com/krishujeniya/sparkwa/actions/runs/42159046566). Captured 16 full-page screenshots across the homepage and VPS-install guide at widths **320, 360, 390, 768, 1024, 1280, 1440, 1920 px**. No measured horizontal overflow at those test sizes.

Browser interactions tested: keyboard search and search results, Escape to close search, saved theme after reload, expanding FAQ, clipboard feedback on code copy, right-side TOC presence, mobile docs sidebar and main menu, Escape behavior, reduced-motion first render, readable guide with JavaScript disabled, visible homepage copy on delayed/failed hero media, and no uncaught page exceptions.

**Screenshot evidence:** [GitHub Actions artifact](https://github.com/krishujeniya/sparkwa/actions/runs/42159046566/artifacts/13474934699).

## Lighthouse (real local lab measurements)

Lighthouse 12.8.2, headless Chrome on GitHub's Ubuntu runner, tested against a locally served production HTML build. Lab results are **not equivalent to real-user Core Web Vitals**; INP is not available as a lab measurement.

| Page and mode | Performance | Accessibility | Best Practices | SEO | LCP | CLS | TBT | FCP |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Homepage, mobile | **99** | **96** | **100** | **100** | 2.0 s | 0 | 0 ms | 0.9 s |
| Homepage, desktop | **100** | **96** | **100** | **100** | 0.4 s | 0.003 | 0 ms | 0.2 s |
| VPS guide, mobile | **100** | **96** | **100** | **100** | 1.4 s | 0 | 0 ms | 0.9 s |
| VPS guide, desktop | **99** | **96** | **100** | **100** | 0.3 s | 0.075 | 0 ms | 0.2 s |

**Lighthouse workflow:** [passed](https://github.com/krishujeniya/sparkwa/actions/runs/42159198825).  
**Raw Lighthouse JSON reports:** [download GitHub Actions artifact](https://github.com/krishujeniya/sparkwa/actions/runs/42159198825/artifacts/13475404446).

## Design brief Section 18 effect coverage

All **93** reference effect IDs documented individually in `docs/DESIGN_SYSTEM.md`: **16 implemented** where useful for documentation and **77 intentionally skipped** with project-specific reasons. Particle engines, animated carousels, commercial pricing widgets, fake demos, and videos without genuine assets were correctly excluded.

## Limitations and follow-up

- Lighthouse accessibility 96/100 is a **lab score, not WCAG 2.2 AA certification**. Manual screen-reader testing, color contrast review in every state and 200% text-zoom checks remain recommended.
- Field LCP / INP / CLS 75th percentile: **not measured** (requires production real-user monitoring, user consent, and traffic).
- Lighthouse was collected for homepage and VPS installation guide only, not each remaining guide.
- Full manual browser Back, non-Chromium browsers, 200% text zoom, third-party outbound URL HTTP checks, and screen-reader narration: **not run**.
- A real video-playback/autoplay audit is **not applicable**, because the repository has no verified product videos and the site has no autoplay video.
- The GitHub runner checked the public site source using a local server; live CDN cache propagation and remote network reliability were not independently tested.
- A limited Markdown renderer produces source-backed HTML for the five allowlisted guides; nonstandard or nested Markdown may render differently than GitHub's Markdown engine.
- The site's documented third-party integration behavior depends on the user's own Google and WhatsApp setup. No external service uptime or official endorsement is implied.
- The one-command installer invokes `sudo`; users should review its contents before execution.

## Maintainer commands

```bash
node scripts/verify-docs.mjs
# Browser QA (needs Playwright and Chrome or Playwright Chromium):
node scripts/browser-docs.mjs
```

No Node dependency is required to serve the site itself; GitHub Pages serves static files from `docs/`.
