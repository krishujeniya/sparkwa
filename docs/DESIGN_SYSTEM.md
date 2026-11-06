# KRISH / LAB — Website Design System

## Intent
An original editorial portfolio for the independent projects of Krish Ujeniya, taking **principles** of cinematic pacing, immersive visuals, intentional typography, and restrained motion from the public Google Flow about page—not Google's marks, images, videos, layout source, or trade dress.

## Brand
- Brand: **KRISH / LAB** — an umbrella label for the portfolio, distinct from the SparkWA product name.
- Original assets: `docs/assets/krish-lab-mark.svg`, `docs/assets/krish-lab-lockup.svg`, and `docs/assets/krish-lab-cover.svg`.
- The abstract K mark consists of intersecting paths on a dark square. SVGs are vector and scale to a favicon.
- Existing products retain their original names, logos, and source identity.

## Palette and type
- Canvas `#060709`; panels `#101218`; muted surfaces `#171a22`.
- Primary text `#f4f5f5`; secondary `#c2c7cf`; muted `#9297a1`.
- Atmosphere: cool blue `#9ab5fc`, with individual product accent colors.
- Display: native sans-serif stack, slightly condensed optical feel via negative letter spacing; serif italic to contrast rhythm.
- Typography relies on native fonts; no blocking font web requests.

## Layout / content
- Fluid editorial hero, cinematic media pane, marquee, brand statement, scroll-snap product exhibition, three-stage narrative, searchable directory, evidence-focused stories, availability, FAQ, footer.
- Product pages share a single JavaScript renderer and a typed JSDoc registry; static HTML entry points allow GitHub Pages hosting without a server.
- Semantic section headings, accessible navigation, form labels, source-grounded media alt text and source links.

## Motion and performance
- Short interaction transitions (roughly 180–250 ms) and larger editorial reveals (up to 750 ms).
- `transform` and `opacity` only for major animations; no WebGL, 3D engine, animation dependency, autoplay video, or external tracking.
- `prefers-reduced-motion` disables the ticker and reveals, and eliminates transforms/scroll animation.
- Native carousel scroll snapping and controls; native details/summary FAQ.

## Asset and provenance policy
- Only verified project media from source repositories; absent files fall back to typographic treatments.
- External media may fail if GitHub is inaccessible; no fabricated mockup substitutes.
- Commercial prices, impact metrics, deployment claims, and customer proof are omitted unless supported in source.

## Accessibility goal
WCAG 2.2 AA is a goal, not independently certified. Inspect focus order, captions, images, interactions, zoom, target size, and contrast before claiming compliance.
