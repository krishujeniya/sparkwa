# KRISH / LAB — QA and Verification Report

Date: 2026-11-06

## Scope
This website is added to an existing Go repository and GitHub Pages deployment. The Go code, CI, release scripts, and original documentation are intentionally preserved. Existing `docs/index.html` is copied verbatim to `docs/sparkwa-docs.html` before the new homepage replaces it.

## Implemented
- Original three-part Krish Lab vector brand: icon/mark, logo lockup, and OG cover.
- Editorial homepage with full-screen hero, media selector, manifesto, native horizontal showcase, interactive three-stage narrative, searchable/filtered 20-entry directory, authentic media gallery, availability, FAQ, and footer.
- Twenty product routes using a shared typed/JSDoc product registry and client-side renderer.
- Distinct product features, audience, source link, provenance, access model, and related work.
- Reduced-motion CSS mode, keyboard focus styles, skip link, native FAQs, labelled buttons, scroll snapping, responsive breakpoints, and media-failure fallbacks.
- No speculative billing, login, forms, user analytics, copied Google media, or autoplay video.

## Checks performed before publishing
- Confirmed the SparkWA repository and the GitHub Pages action (publishes `docs/` on changes).
- Read the repository tree, README, existing HTML, and sitemap.
- Inspected the README files of the 20 selected product projects and checked image asset paths against source repository trees where possible.
- Confirmed the existing SparkWA documentation content is retained in a separate HTML file.
- Created all product route entries and shared source files as Git blobs prior to commit.

## Not independently verified at authoring time
- Browser-based viewport tests at 360/390/768/1024/1280/1440/1920: **not yet run**.
- Lighthouse CWV targets (LCP ≤2.5s, INP ≤200ms, CLS ≤0.1): **not measured**.
- Accessibility audit / WCAG 2.2 AA: **not certified**.
- Runtime link checks for third-party websites and external media: **not exhaustive**.
- Go test/build status: **not claimed** unless a post-commit CI run confirms it.
- GitHub Pages production deployment status is separate from the Git commit status.

## Manual follow-up checklist
1. Confirm Pages workflow deploys the new commit and visit the production URL.
2. Inspect each viewport and every `products/<slug>/` route.
3. Tab through menu, showcase, stage selectors, filters, search, and FAQ.
4. Toggle OS reduced-motion preferences; review under 200%/400% browser zoom.
5. Check every external image URL for successful loads and correct attribution.
6. Run Lighthouse mobile/desktop and inspect browser console and resource waterfall.
7. Confirm the original SparkWA documentation is fully usable on `/sparkwa/sparkwa-docs.html`.
