# Visual reference

Values measured on November 1, 2026 with headless Chromium (Playwright), 1440×900 viewport, from the computed styles of:

- `platform.claude.com/docs` (the `docs` layout)
- `www.claude.com` and `www.anthropic.com` (the `landing` layout)
- `www.anthropic.com/research` and `/news` articles (the `article` layout)

The measurement scripts are in `tools/visual/`. Measure again before changing tokens: the sites change.

## Typography

All three sites use their own fonts (Anthropic Sans, Anthropic Serif, Anthropic Mono). They are proprietary and cannot be embedded. We compared open-licensed candidates side by side with screenshots:

| Role | Choice (OFL) | Why | Rejected |
|---|---|---|---|
| Sans | Hanken Grotesk | The closest grotesque in metrics and rhythm; at 832 px the body paragraph breaks its lines at the same points as the reference | Inter, Geist, Onest and Public Sans (too wide), Manrope (too geometric). Instrument Sans and Schibsted Grotesk did not load in the test and were not evaluated |
| Serif | Newsreader | A contemporary text serif with an `opsz` axis; contrast and terminals are close at 26 px and at 56 px | Source Serif 4 (heavy at display sizes), Literata and Noto Serif (wide), Gelasio, Spectral and Caslon (too classical) |
| Mono | JetBrains Mono | anthropic.com itself loads JetBrains Mono as a code font | — |

Licenses are in `internal/theme/assets/fonts/OFL-*.txt`.

## `docs` layout (platform.claude.com/docs)

| Element | Light | Dark | Metrics |
|---|---|---|---|
| Page background | `#fcfcfb` | `#151515` | |
| Strong text (headings, active nav) | `#0b0b0b` | `#f0efec` | |
| Body text | `#52514e` | `#c3c2b7` | sans 15px / 22px |
| Muted text (breadcrumb, TOC, inactive nav) | `#898781` | `#898781` | 14px / 20px |
| Borders | `rgba(11,11,11,.1)` | `rgba(255,255,255,.1)` | 1px |
| Raised surface (active, th) | `rgba(11,11,11,.05–.1)` | `rgba(255,255,255,.05–.15)` | |
| Link | `#184f95`, underline `rgba(24,79,149,.4)`, offset 3px | `#6da7ec` | |
| Inline code | `#8e2626` on `rgba(11,11,11,.05)`, border `.1` | `#ec7e7e` on `rgba(255,255,255,.05)` | mono 13.5px, radius 5.4px, padding 0.84px 3.4px |
| Code block | background `#fff`, border `.1`, radius 11px | background `#1a1a19` | mono 13px / 19px, tabbed header 14px weight 500 |
| Syntax highlighting | GitHub Primer palette: text `#1f2328`, keyword `#cf222e`, string `#0a3069`, number/builtin `#0550ae`, parameter `#953800` | VS Code Dark+ style: `#9cdcfe`, `#ce9178`, `#b5cea8`, `#dcdcaa` | |

Structure:

- Header: 56px tall, bottom border, 16px horizontal padding. Serif logo around 20px. Top navigation with 32px pill items; the active item has a `rgba(11,11,11,.1)` background.
- Sidebar: 264px, right border. 32px items, 8px radius, 8px horizontal padding, 14px. The active item has a `rgba(11,11,11,.1)` background and strong text. Group label: 14px weight 500, strong text.
- Content: starts 45px after the sidebar, 832px wide.
  - Breadcrumb: 14px, muted.
  - H1: serif 26.25px weight 520, line height 1.25.
  - Lead: sans 16.875px / 1.5, body color.
  - H2: serif 20.625px weight 520. H3: serif 16.875px weight 520.
  - Strong: weight 600, strong text.
- TOC: 224px on the right, 32px vertical padding. 28px items with 12px left padding. Active: weight 500, strong text and a 2px left border. Level 3 indented by 12px.
- Table: 1px border, 7px radius. th with a `.05` background, weight 500, 6.5px 11.25px padding.
- Callout: 11px radius, 10px 14px padding, 14px / 20px, icon on the left.
- Home cards: 1px border, about 12px radius, `#fff` background.

## `landing` layout (claude.com / anthropic.com)

| Element | Value |
|---|---|
| Background | `#faf9f5` (ivory). Dark on claude.com: `#141413`, text `#faf9f5` |
| Text | `#141413`; secondary `#5e5d59` |
| Hero H1 | serif 56px weight 400, line height 1.15 |
| Hero subtitle | serif 18px |
| Section H2 | serif 52px weight 500 |
| Primary button | background `#141413`, text `#faf9f5`, radius 8px, 15px weight 500 |
| Secondary button | 1px `#d1cfc5` border, transparent background, radius 8px |
| Card | background `#fff` (dark: `#1f1e1d`), border `#e8e6dc`, radius 16–32px |
| Accent | clay `#d97757`, interactive `#c6613f` |

## Warm gray scale (anthropic.com and claude.com)

`000 #fff`, `050 #faf9f5`, `100 #f5f4ed`, `150 #f0eee6`, `200 #e8e6dc`, `250 #dedcd1`, `300 #d1cfc5`, `350 #c2c0b6`, `400 #b0aea5`, `450 #9c9a92`, `500 #87867f`, `550 #73726c`, `600 #5e5d59`, `650 #4d4c48`, `700 #3d3d3a`, `750 #30302e`, `800 #262624`, `850 #1f1e1d`, `900 #1a1918`, `950 #141413`, `1000 #000`.

Accents: clay `#d97757` (interactive `#c6613f`), olive `#788c5d`, sky `#6a9bcc`, fig `#c46686`, cactus `#bcd1ca`, heather `#cbcadb`, oat `#e3dacc`, error `#bf4d43`.

Spacing: 4, 8, 12, 16, 24, 32, 40, 48, 64, 80, 96, 128. Radii: 4, 8, 12, 16. Easing: `cubic-bezier(.165,.84,.44,1)` (out-quart).

## `article` layout (anthropic.com/research and /news)

Measured on `anthropic.com/research/claude-think-tool` and `anthropic.com/news/2026-usage-policy-update`:

- Centered title: sans weight 700, 32.5px; a 32px serif subtitle right below.
- Three columns at 1440px: numbered chapters (`01`, `02`… in 12px mono) at x=80, 198px wide, with borders between items; a 625px body centered at x=400; facts (Published, Team, Topics) at x=1148, 197px wide.
- Body: serif 16.7px / 25.5px (news) or sans (research), color `#141413`. H2 sans 700 26px, 48px above and 24px below. Paragraphs 24px apart. Links underlined in the text color.

## Motion (anthropic.com, claude.com, platform.claude.com)

From the `@keyframes` and `transition` rules in the three sites' CSS:

- Curves: `cubic-bezier(.165,.84,.44,1)` (out-quart) and `cubic-bezier(.19,1,.22,1)` (out-expo).
- Durations: 0.15 to 0.3s for hover and color; 0.5 to 0.8s for entrances.
- Entrances: `word-rise` (translateY .42em), `unblur` (blur 4px → 0) and `fade-in` on anthropic.com; `raise` (translateY 16px) and `fadeUp` on claude.com; `reveal-in` (translateY 4px) in the docs.
- The docs use view transitions and `animation-timeline`; anthropic.com animates strokes with `stroke-dash`.
- Nothing flickers or loops in reading content.

## Illustrations

The icons on platform.claude.com/docs and claude.com are thick black strokes (about 2px at 48px), slightly irregular corners and a flat warm fill (`#e3dacc` oat, `#dedcd1`). The large pieces on anthropic.com are engravings and collages with texture. Wavemark generates its own art with the same vocabulary (a pen stroke and an oat fill out of register) without copying any piece.
