---
name: Pregunta
colors:
  canvas: '#f7f8fb'
  surface: '#ffffff'
  surface-2: '#f6f8fb'
  surface-footer: '#fbfcfd'
  chip: '#f2f4f8'
  panel: '#eff4ff'
  line: '#e3e7ef'
  line-2: '#d5dbe6'
  line-soft: '#eef1f6'
  line-hover: '#c9d3e6'
  ink: '#10182b'
  ink-body: '#2b3242'
  ink-muted: '#45464c'
  on-dark: '#c6cfdd'
  dark-2: '#1c2742'
  accent-on-dark: '#7aa2ff'
  primary: '#2563eb'
  agent-cyan: '#38bdf8'
  amber: '#d97706'
  amber-text: '#b45309'
  amber-bg: '#fff7ea'
  amber-border: '#fde68a'
  status-vigente: '#10b981'
  status-vigente-text: '#065f46'
  status-vigente-bg: '#ecfdf5'
  status-vigente-border: '#a7f3d0'
  status-derogada: '#ef4444'
  status-derogada-text: '#991b1b'
  status-derogada-bg: '#fef2f2'
  status-derogada-border: '#fecaca'
  code-bg: '#10182b'
  code-button: '#2a3756'
  code-prompt: '#6ee7b7'
  code-text: '#e6ebf5'
typography:
  display:
    fontFamily: Bricolage Grotesque
    fontSize: clamp(36px, 6vw, 66px)
    fontWeight: '700'
    lineHeight: '1.02'
    letterSpacing: -0.03em
  headline-section:
    fontFamily: Bricolage Grotesque
    fontSize: clamp(30px, 4.4vw, 46px)
    fontWeight: '700'
    lineHeight: '1.05'
    letterSpacing: -0.025em
  headline-cta:
    fontFamily: Bricolage Grotesque
    fontSize: clamp(34px, 5vw, 54px)
    fontWeight: '700'
    lineHeight: '1.02'
    letterSpacing: -0.03em
  headline-card:
    fontFamily: Figtree
    fontSize: 1.25rem
    fontWeight: '700'
    lineHeight: '1.25'
  lead:
    fontFamily: Figtree
    fontSize: clamp(17px, 2vw, 20px)
    fontWeight: '400'
    lineHeight: '1.6'
  body:
    fontFamily: Figtree
    fontSize: 1rem
    fontWeight: '400'
    lineHeight: '1.6'
  quote:
    fontFamily: Figtree
    fontSize: 1.1875rem
    fontWeight: '400'
    lineHeight: '1.6'
  section-label:
    fontFamily: Figtree
    fontSize: 0.875rem
    fontWeight: '600'
  label-caps:
    fontFamily: Figtree
    fontSize: 0.8125rem
    fontWeight: '600'
    letterSpacing: 0.06em
  code:
    fontFamily: JetBrains Mono
    fontSize: 0.8125rem
    fontWeight: '400'
    lineHeight: '1.6'
rounded:
  mark: 4px
  code: 14px
  card-small: 18px
  card: 24px
  card-guarantee: 20px
  cta: 32px
  full: 999px
spacing:
  page-max: 1280px
  reading-max: 960px
  margin: clamp(20px, 4.4vw, 56px)
  section: clamp(64px, 9vw, 112px)
  section-first: clamp(56px, 7vw, 88px)
  card: clamp(22px, 3vw, 32px)
  card-wide: clamp(24px, 3vw, 36px)
---

> **En kitlegal.es** (ADR 0024) estos tokens están en `web/src/styles/global.css`, con nombres en español, y las clases que se repiten, en `web/src/lib/clases.ts`. Es la dirección «Pregunta» del rediseño de octubre de 2026: la portada empieza por la duda de quien pregunta, con tarjetas blancas de borde fino, botones en pastilla y un azul primario. Los colores de estado y el ámbar tienen un tono de texto (`-text`) que alcanza el contraste 4,5:1 del nivel AA; el tono vivo es para puntos, bordes y rellenos. La página no ejecuta JavaScript (su política de seguridad lo prohíbe): las pestañas, el menú móvil y el selector de sistema son botones de opción y `<details>` con CSS.

## Brand & Style

The site speaks to the people who use kitlegal, not to engineers (ADR 0034): a citizen with a deadline, a lease or a job, and a law firm. Someone with no technical background should understand in ten seconds what kitlegal does, trust its rigour and install it. The governing premise is still **"Nada sin cita"** (Nothing without citation): every statement about a norm shows the literal BOE text that supports it, with its identifier, block, consultation date and hash.

### Visual Style
1. **The question first**: the home page opens with the reader's doubt and an interactive box of example questions, each answered with the literal BOE text.
2. **Clear and close**: a light canvas, white cards with hairline borders, pill-shaped buttons and chips, and a confident grotesque for headlines.
3. **Rigour you can see**: validity chips, the "with and without kitlegal" comparison and the monospaced provenance under every quote.

## Colors

The palette is light and calm, so that the legal text and its provenance carry the weight.

- **Canvas (`#f7f8fb`) and white cards** with a `#e3e7ef` hairline. Inner separators use `#eef1f6`; the quote column of a query point sits on `#f6f8fb`, card footers on `#fbfcfd` and category chips on `#f2f4f8`.
- **Ink (`#10182b`)** for text, dark buttons and the dark "why trust it" cards, where secondary text is `#c6cfdd` and numbers are `#7aa2ff`.
- **Primary blue (`#2563eb`)** for main buttons, links, section labels and the final call to action; `#eff4ff` for the "short answer" chip, the active navigation item and question bubbles.
- **Statutory states**, unchanged in meaning: *vigente* green (`#10b981`, text `#065f46`, background `#ecfdf5`, border `#a7f3d0`, which also highlights the key words of a quote), *derogada* red (`#ef4444`, `#991b1b`, `#fef2f2`, `#fecaca`) and the amber warning (`#d97706`, text `#b45309`, background `#fff7ea`, border `#fde68a`) for "texto posiblemente desactualizado", the limits of a page and what is still untested.
- **Dots** mark guarantee cards and categories: blue for citation, green for validity, amber for price or scope, cyan (`#38bdf8`) for privacy or licence.
- **Code blocks**: `#10182b` background, `#6ee7b7` prompt, `#e6ebf5` text.

## Typography

1. **Headlines (Bricolage Grotesque, 700)**: page titles, section titles, numbers and the logo, with tight tracking (`-0.03em`) and optical sizing.
2. **Text and interface (Figtree, 400–700)**: body copy, navigation, buttons, cards and the literal text of the norms, which is quoted in Figtree at 15–21px, never in italics.
3. **Machine-verifiable data (JetBrains Mono)**: BOE identifiers (`[BOE-A-2015-10565, bloque a21]`), consultation dates, hashes and commands.

Microcopy never drops below 12px. Section labels are 14px semibold in the primary blue; uppercase labels are 13px semibold with `0.06em` tracking.

## Layout & Spacing

Content is centred at a maximum of 1280px with a `clamp(20px, 4.4vw, 56px)` side margin; reading columns (the question box, a query page, the install guide) stop at 880–960px. Sections are separated by `clamp(64px, 9vw, 112px)`, and the first one after the hero by `clamp(56px, 7vw, 88px)`. Cards are padded `clamp(22px, 3vw, 32px)` and grids leave 16–20px gaps.

Every grid is `repeat(auto-fit | auto-fill, minmax(min(Npx, 100%), 1fr))`, so it falls to one column on a phone without media queries, and type sizes use `clamp()`. The only explicit breakpoint is the header's, at 860px, below which the navigation folds into a "Menú" button. Touch targets are at least 44px tall.

## Shapes

Rounded, friendly and clear: the pill (`999px`) is now the shape of every button, chip, tab and the navigation bar. Large cards use `24px`, the guarantee cards and the mobile menu `20px`, the final call to action `32px`, sub-cards and code blocks `14–18px`, and highlighted text `4px`. Depth comes from 1px borders; only the featured question card carries a soft shadow (`0 20px 50px -20px rgba(16,24,43,.18)`), and query cards gain `0 20px 50px -24px rgba(16,24,43,.28)` with a `#c9d3e6` border on hover. Links and buttons drop to `opacity: .8` on hover.

## Components

- **Header**: the logo (32px icon and "kitlegal" in Bricolage 24/700), the navigation in a white pill (`5px` padding, links `8px 14px`, the current one on `#eff4ff` in blue) and the "Instalar gratis" button.
- **Buttons**: primary blue with white text (`15px 26px`, 16/600), secondary white with a `#d5dbe6` border, dark `#10182b`; on the blue call to action, white or white-outlined.
- **Status chip**: 14/600, `6px 14px`, in the *vigente* colours.
- **Question box**: four topic chips (the active one filled with ink) above a white card with the question, the "Texto literal del BOE" chip, the quote, its reference in mono and the link to the full answer.
- **Comparison**: two cards, "sin kitlegal" on the *derogada* background with the errors highlighted, and "con kitlegal" on the *vigente* background with the correct figure highlighted and the reference in a white box.
- **Steps**: white pills with a 28px numbered circle.
- **Install steps**: numbered 48px circles, download buttons with an icon, the red-warning explanation in amber, and command blocks that select the whole line on click.
