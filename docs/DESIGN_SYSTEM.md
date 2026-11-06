# SparkWA Documentation Design System

## Scope and source confidence
- **Site type:** single-product, open-source developer documentation with a short product overview; not an AI marketplace or portfolio.
- **Audience:** Ubuntu/Debian/macOS operators, developers, and Gemini Spark users who need setup, OAuth, troubleshooting and security information.
- **Theme:** light-first editorial documentation with user-toggleable dark reading theme. This is a project-specific choice recommended by the attached reference brief, not a reconstruction of Google’s CSS.
- **Hero style:** source-media-led, using the real SparkWA repository banner. **No fabricated footage, demos or product UIs.**
- **Evidence labels:** The attached brief reports Antigravity and Flow observations/design confirmations. Those labels are retained in the table; underlying websites and design source files were **not independently confirmed** for this audit.
- **Existing functionality:** The Go executable, OAuth, Gmail, WhatsApp and service logic are untouched. The original single-page docs are at `reference/complete-guide.html`.

## 2026-11-06 responsive header and motion refinement
- Fixed mobile navigation appearing as an unstyled duplicate row at desktop widths by providing explicit default hidden styling, a >=901px hard hide and a breakpoint-specific animated mobile drawer. The navigation still works with the native link list and is controlled by the menu button.
- Refined hero typography into two intentional lines ("Your agent." / "On WhatsApp.") for desktop and small screen widths.
- Implemented meaningful, small animations: hover and focus states on header links, cards and workflow steps; an ambient original CSS ring/aura on the real project artwork; one-shot scroll reveals on homepage editorial components; a progress indicator and a scrolled header surface.
- Animations default to disabled for `prefers-reduced-motion: reduce`; the site keeps all technical guide bodies visible and uses no autoplay media. The original SparkWA branding and documentation content stay intact.
- Responsive navigation uses 900px as the compact-header breakpoint; the mobile drawer closes when resizing to desktop or tapping outside it.

## Visual tokens
The source of truth is `docs/assets/docs.css`: paper `#fcfbf9`, white `#fff`, dark ink `#161719`, soft surface `#f5f4f0`, readable muted text `#60656c`, spark-orange accent `#d96036`, black terminal `#131416`, gentle borders `#e6e5e2`. User-chosen dark mode uses matching contrast-oriented tokens. Preferred font stack is system sans + system monospace, no copied/private Google fonts.

- Grid: max width 1420px; guide content capped to 760px; 255px docs sidebar and optional 210px TOC. On narrow viewports the TOC is removed and the sidebar becomes a drawer.
- Responsive breakpoints: 1250, 900, 720, 380px.
- Keyboard: skip navigation, Ctrl/Cmd+K and slash for search, Escape to close, normal native tab order.
- Code: dark accessible code blocks, horizontal scrolling, copy controls with result announcement.
- Native anchor navigation, stable print layout, no auto-play video, no particle interference behind text.
- Logo: existing SparkWA bridge mark `docs/assets/sparkwa-mark.svg`; repository banner `docs/assets/sparkwa-banner.png` retained.
- Intro motion only: CSS opacity/translate and hero art diffusion; `prefers-reduced-motion` disables animated movement.
- Search dialog: keyboard focus loop, background close, result ranking by title/description/body and empty state.
- Reliability: static HTML article bodies and guides still work without JavaScript; search, theme toggle, code copy enhance rather than gate reading.

## Verified and unverified
This document records **implementation decisions**, not independently audited WCAG certification or field Web Vitals. Human browser checks and Lighthouse need actual execution; consult `docs/QA_REPORT.md`.

## Section 18 effects coverage — full row-by-row audit
The attached prompt defines **93 effect IDs**. Implemented 16 applicable patterns and intentionally skipped 77, generally because a single-product docs site needs information retrieval more than expensive motion. **Built** means source code was implemented; it does not mean the behavior has passed a live visual check.

| ID | Effect | Trigger | Properties | Duration | Easing | Reduced motion | Status / skip reason | Evidence | Where |
|---|---|---|---|---|---|---|---|---|---|
| L1 | Headline entrance | page load | opacity,translateY | 700ms | expo | static text | **Built** | O | Hero headline on homepage |
| L2 | Subcopy and CTA follow-up | page load | opacity,translateY | 700ms | expo | static text | **Built** | R | Homepage hero subcopy/actions |
| L3 | Hero media diffusion | page load | opacity,scale,blur | 750ms | expo | static media | **Built** | D | Homepage project artwork |
| L4 | Particle field fade | page load | canvas opacity | 700ms | primary | none | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| L5 | Announcement entrance | page load | opacity,height | 250ms | primary | static banner | Skipped: Announcement remains visible; no entrance delay | O | — |
| L6 | Header reveal | page load | opacity,translateY | 250ms | primary | static header | Skipped: Header is immediately available for reading | R | — |
| L7 | Scroll cue loop | idle | translateY,opacity | 1800ms | ease-in-out | static cue | Skipped: No distracting loop next to documentation | O | — |
| A1 | Particle drift | visible | position,rotation | continuous | physics | stopped | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| A2 | Particle turbulence | visible | velocity | continuous | physics | stopped | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| A3 | Particle orbital flow | mode toggle | target position | continuous | spring | static | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| A4 | Halos and waves | mode change | target position | continuous | spring | static | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| A5 | Floating chip wave | idle or scroll | transform | 7s | sine | static | Skipped: Particle and cursor physics excluded for reading focus and performance | R | — |
| A6 | Atmospheric glow breathing | idle | opacity,scale | 8s | ease-in-out | static glow | Skipped: Particle and cursor physics excluded for reading focus and performance | R | — |
| A7 | Ambient video loop | visible | video frames | loop | media | poster | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| A8 | Content color emergence | media change | accent | 450ms | primary | instant | Skipped: Particle and cursor physics excluded for reading focus and performance | D | — |
| P1 | Particle pointer repulsion | pointer | position | continuous | physics | disabled | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| P2 | Custom cursor | pointer | transform | 350ms | quickTo | disabled | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| P3 | Magnetic buttons | pointer | translate | 250ms | primary | disabled | Skipped: Particle and cursor physics excluded for reading focus and performance | R | — |
| P4 | Magnetic border | pointer | border gradient | 150ms | linear | solid border | Skipped: Particle and cursor physics excluded for reading focus and performance | R | — |
| P5 | Button hover lift | hover | translateY,background | 150ms | primary | color feedback | **Built** | I | Primary and secondary CTAs |
| P6 | Button press scale | press | scale | 90ms | primary | color feedback | Skipped: Buttons use native :active states without scaling | R | — |
| P7 | Animated underline | hover | scaleX | 200ms | primary | underline | Skipped: Links use persistent underlines rather than motion | R | — |
| P8 | Nav hover color | hover | color | 200ms | primary | instant color | **Built** | I | Header navigation links |
| P9 | Chevron rotate | open | rotate | 200ms | primary | instant | Skipped: Nonessential decorative effect intentionally omitted for documentation | R | — |
| P10 | Card hover emphasis | hover | translateY,border | 250ms | primary | border | **Built** | R | Doc cards and process steps |
| P11 | Directory preview swap | hover | preview image | 250ms | primary | instant | Skipped: Nonessential decorative effect intentionally omitted for documentation | O | — |
| P12 | Focus ring | focus-visible | outline | instant | none | unchanged | **Built** | R | Interactive controls across docs |
| P13 | Variable-font shift | hover | font-weight | 250ms | primary | none | Skipped: Nonessential decorative effect intentionally omitted for documentation | R | — |
| P14 | Video hover preview | hover | video | 250ms | fade | poster | Skipped: Nonessential decorative effect intentionally omitted for documentation | R | — |
| S1 | Scroll reveal | enter view | opacity,translateY | 450ms | reveal | visible | Skipped: Reading text must not start hidden | R | — |
| S2 | Group stagger | enter view | opacity,translateY | 70ms step | reveal | visible | Skipped: No staged hiding of reading content | R | — |
| S3 | Heading mask | enter view | clip,transform | 600ms | expo | visible | Skipped: Heading text stays indexable and readable | R | — |
| S4 | Media scroll entrance | enter view | opacity,scale,blur | 700ms | expo | visible | Skipped: Content never waits for scroll | D | — |
| S5 | Parallax media | scroll | translateY | scrub | linear | none | Skipped: No scroll hijacking or page-visibility-dependent content in docs | R | — |
| S6 | Preview parallax | scroll | translateY | scrub | linear | none | Skipped: No scroll hijacking or page-visibility-dependent content in docs | O | — |
| S7 | Scrubbed objects | scroll | transform | scrub | linear | static | Skipped: No scroll hijacking or page-visibility-dependent content in docs | R | — |
| S8 | Sine-wave accent row | scroll | translate | scrub | sine | static | Skipped: No scroll hijacking or page-visibility-dependent content in docs | R | — |
| S9 | Pinned stage | scroll | pin,crossfade | scrub | linear | stack | Skipped: No scroll hijacking or page-visibility-dependent content in docs | R | — |
| S10 | Scroll progress | scroll | scaleX | scrub | linear | hidden | Skipped: No scroll hijacking or page-visibility-dependent content in docs | R | — |
| S11 | Active nav indicator | scroll | color | 250ms | primary | instant | Skipped: TOC has a separate active state (S16) | O | — |
| S12 | Header scroll surface | scroll | background | 250ms | primary | instant | Skipped: Stable sticky reading header | R | — |
| S13 | Scroll cue fade | scroll | opacity | 250ms | primary | hidden | Skipped: No scroll hijacking or page-visibility-dependent content in docs | R | — |
| S14 | Smooth scrolling override | wheel | scroll position | 600ms | smoother | native | Skipped: No scroll hijacking or page-visibility-dependent content in docs | R | — |
| S15 | Anchor navigation | anchor click | scroll position | native | browser | instant | **Built** | R | Native anchors, scroll-padding |
| S16 | Docs active TOC | scroll | aria-current,link color | instant | native | instant | **Built** | R | Doc article right-side TOC |
| C1 | Product carousel | arrows,swipe | translateX | 500ms | ease-in-out | snap | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C2 | Active card emphasis | slide | scale,opacity | 450ms | primary | opacity | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C3 | Peeking neighbors | idle | edge mask | static | none | static | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C4 | Carousel arrow states | hover | fill,opacity | 150ms | primary | color | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C5 | Pagination fill | slide | width | 450ms | primary | instant | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | R | — |
| C6 | Auto rotation | timer | slide | 6s | ease-in-out | off | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | R | — |
| C7 | Stage indicator | select | translateX | 250ms | primary | instant | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C8 | Stage crossfade | select | opacity,translateY | 450ms | reveal | instant | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C9 | Stage word transition | select | mask | 600ms | expo | instant | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C10 | Use-case switcher | select | opacity | 250ms | primary | instant | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | R | — |
| C11 | Story selector | select | media | 450ms | primary | instant | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | O | — |
| C12 | Pricing card entrance | enter view | opacity,border | 450ms | primary | visible | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | R | — |
| C13 | Accordion animation | click | grid rows | 250ms | primary | instant | Skipped: Native details open instantly to prioritize reading | O | — |
| C14 | Code/docs tabs | select | indicator | 200ms | primary | instant | Skipped: Single-product documentation does not need a carousel, stage or pricing selector | R | — |
| O1 | Mobile menu overlay | tap | display,panel | instant | native | instant | **Built** | O | Mobile navigation and sidebar (instant by design) |
| O2 | Docs search dialog | Ctrl/Cmd+K | opacity,translate,scale | 220ms | expo | instant | **Built** | R | Keyboard search overlay |
| O3 | Video modal | click | opacity,scale | 250ms | primary | instant | Skipped: No verified video asset exists | O | — |
| O4 | Dropdown panel | open | opacity,translate | 200ms | primary | instant | Skipped: No dropdown necessary | O | — |
| O5 | Copy feedback label | success | text | immediate + 1700ms hold | native | instant | **Built** | O | Copy code / logo success feedback |
| O6 | Tooltip | focus,hover | opacity | 120ms | primary | instant | Skipped: Nonessential decorative effect intentionally omitted for documentation | R | — |
| M1 | Video poster crossfade | play | opacity | 250ms | primary | poster | Skipped: No verified playable video asset in repository | O | — |
| M2 | Custom control auto-hide | idle | opacity | 200ms | primary | visible | Skipped: No verified playable video asset in repository | R | — |
| M3 | Play control hover | hover | scale | 150ms | primary | color | Skipped: No verified playable video asset in repository | O | — |
| M4 | Offscreen video pause | visibility | media state | instant | native | poster | Skipped: No verified playable video asset in repository | R | — |
| M5 | Image blur-up | page load | blur,opacity | 750ms | expo | static | **Built** | D | Hero media initial diffusion |
| M6 | Image fade-in | load | opacity | 200ms | primary | instant | Skipped: Image retains layout; no load-dependent hiding | R | — |
| T1 | Per-character stagger | view | opacity,transform | 22ms step | reveal | static | Skipped: Static factual content remains readable without animation | O | — |
| T2 | Whole-line reveal | load | opacity,translateY | 700ms | expo | static | **Built** | O | Headline and hero entrance |
| T3 | Typewriter | view | text,cursor | 40ms/char | steps | full text | Skipped: Static factual content remains readable without animation | R | — |
| T4 | Count-up metrics | view | number | 900ms | expo | final | Skipped: Static factual content remains readable without animation | R | — |
| T5 | Async status text | async | text | instant | native | final text | Skipped: No real asynchronous application work on site | O | — |
| R1 | View-transition crossfade | route | opacity | 180ms | native | none | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R2 | Shared-element morph | route | transform | 350ms | expo | none | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R3 | Persistent route header | route | held DOM | static | native | reload | Skipped: Static HTML page navigation intentionally uses full reload | R | — |
| R4 | Chapter preview | section | opacity,translate | 700ms | expo | static | Skipped: Not a dynamic SaaS application; static document navigation favored | D | — |
| R5 | Explorer reflow | search | translate | 300ms | primary | instant | Skipped: Not a dynamic SaaS application; static document navigation favored | O | — |
| R6 | Explorer empty | search | opacity | 200ms | primary | instant | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R7 | Skeleton shimmer | loading | background | 1400ms | linear | static | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R8 | Button spinner | loading | rotate | 800ms | linear | static | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R9 | Form success check | submit | stroke | 400ms | primary | static | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R10 | Form error | submit | border | 200ms | primary | static | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R11 | Toggle control | change | background,indicator | 150ms | primary | instant | Skipped: Not a dynamic SaaS application; static document navigation favored | R | — |
| R12 | Particle mode toggle | change | spring | 250ms | spring | static | Skipped: Particle and cursor physics excluded for reading focus and performance | O | — |
| R13 | Theme toggle | click | color tokens | instant | native | instant | **Built** | R | Theme toggle with saved preference |
| R14 | Demo replay controls | click | timeline | 250ms | primary | step | Skipped: Not a dynamic SaaS application; static document navigation favored | O | — |
| R15 | SVG stroke draw | view | stroke-dashoffset | 800ms | primary | drawn | Skipped: Diagram is static for comprehension | R | — |
| R16 | Copy-logo label | success | text | immediate+1700ms | native | instant | **Built** | O | Copy logo SVG control |
| R17 | Language/download expansion | click | height,opacity | 250ms | primary | instant | Skipped: No translated pages or complex download selector | O | — |

## Styles and states coverage
Built: neutral surfaces, 1px borders, 9–32px radius system, pill action controls, docs nav and TOC, search input/results, themed code blocks, copy buttons, responsive table wrappers, active current page, focus rings, dynamic success announcement, reading-side print styles, optional dark mode, and screen-reader-aware search dialog.

Skipped: pricing cards, fake SaaS form states, custom cursor, three-dimensional or particle physics, video playback controls without verified footage, language dropdown without translations, invented testimonials or status metrics, marketing carousels and product registries. CSS hover is never the sole path to an action.

Reduced motion: stop hero/load transforms, retain static images and text, native anchor navigation, native accordion and code copy state, no ambient continuous animation.

**Counts: 16 built patterns; 77 deliberately skipped; 93 audited.**
