---
name: Jurisprudence & Code
colors:
  canvas: '#f8f9ff'
  panel: '#eff4ff'
  panel-strong: '#e5eeff'
  ink: '#0b1c30'
  ink-muted: '#45464c'
  slate-midnight: '#0B1325'
  slate-surface: '#0F172A'
  amber-seal: '#D97706'
  amber-parchment: '#B45309'
  agent-cyan: '#38BDF8'
  agent-blue: '#2563EB'
  status-vigente: '#10B981'
  status-vigente-text: '#065F46'
  status-vigente-bg: '#ECFDF5'
  status-vigente-border: '#A7F3D0'
  status-agotada: '#F59E0B'
  status-agotada-text: '#92400E'
  status-agotada-bg: '#FFFBEB'
  status-agotada-border: '#FDE68A'
  status-derogada: '#EF4444'
  status-derogada-text: '#991B1B'
  status-derogada-bg: '#FEF2F2'
  status-derogada-border: '#FECACA'
  boe-slate-border: '#E2E8F0'
  canvas-ivory: '#F8FAFC'
  code-dark-bg: '#0A0E17'
  code-dark-border: '#1E293B'
typography:
  display:
    fontFamily: Newsreader
    fontSize: 3.25rem
    fontWeight: '600'
    lineHeight: 3.75rem
    letterSpacing: -0.025em
  headline-lg:
    fontFamily: Newsreader
    fontSize: 2.25rem
    fontWeight: '600'
    lineHeight: 2.75rem
    letterSpacing: -0.02em
  headline-lg-mobile:
    fontFamily: Newsreader
    fontSize: 1.75rem
    fontWeight: '600'
    lineHeight: 2.25rem
    letterSpacing: -0.015em
  headline-md:
    fontFamily: Newsreader
    fontSize: 1.75rem
    fontWeight: '600'
    lineHeight: 2.25rem
    letterSpacing: -0.015em
  headline-sm:
    fontFamily: Newsreader
    fontSize: 1.25rem
    fontWeight: '600'
    lineHeight: 1.75rem
    letterSpacing: -0.01em
  body-lg:
    fontFamily: Inter
    fontSize: 1.125rem
    fontWeight: '400'
    lineHeight: 1.875rem
  body-md:
    fontFamily: Inter
    fontSize: 0.9375rem
    fontWeight: '400'
    lineHeight: 1.625rem
  body-sm:
    fontFamily: Inter
    fontSize: 0.8125rem
    fontWeight: '400'
    lineHeight: 1.375rem
  legal-article:
    fontFamily: Newsreader
    fontSize: 1.0625rem
    fontWeight: '400'
    lineHeight: 1.75rem
    letterSpacing: 0.005em
  code-base:
    fontFamily: JetBrains Mono
    fontSize: 0.875rem
    fontWeight: '400'
    lineHeight: 1.5rem
  code-sm:
    fontFamily: JetBrains Mono
    fontSize: 0.75rem
    fontWeight: '400'
    lineHeight: 1.25rem
  citation-stamp:
    fontFamily: JetBrains Mono
    fontSize: 0.6875rem
    fontWeight: '500'
    lineHeight: 1rem
    letterSpacing: 0.06em
  label-md:
    fontFamily: Inter
    fontSize: 0.875rem
    fontWeight: '500'
    lineHeight: 1.25rem
    letterSpacing: 0.01em
  label-sm:
    fontFamily: Inter
    fontSize: 0.75rem
    fontWeight: '600'
    lineHeight: 1rem
    letterSpacing: 0.03em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  gutter: 1.5rem
  gutter-mobile: 1rem
  margin: 2rem
  margin-mobile: 1rem
  space-2xs: 0.125rem
  space-xs: 0.25rem
  space-sm: 0.5rem
  space-md: 1rem
  space-lg: 1.5rem
  space-xl: 2rem
  space-2xl: 3rem
  space-3xl: 4.5rem
---

> **En kitlegal.es** (ADR 0024) estos tokens están en `web/src/styles/global.css`, con nombres en español. Los colores de estado y el ámbar tienen un tono de texto (`-text`, y `amber-parchment` para el ámbar) que alcanza el contraste 4,5:1 del nivel AA; el tono vivo es para iconos, bordes y rellenos. De la paleta generada que traía la exportación de Stitch solo quedan los neutros que usa la portada (`canvas`, `panel`, `panel-strong`, `ink`, `ink-muted`); el resto (`primary: #000000`, `secondary`, `tertiary`…) se ha quitado porque contradecía la marca.

## Brand & Style

This design system occupies the intersection between solemn legislative institutions and developer-first utility tooling. It is conceived for environments where artificial intelligence agents (such as MCP runtimes, Claude Code, and terminal daemons) consult and operationalize primary legal sources. The governing premise is **"Nada sin cita"** (Nothing without citation): absolute legal certitude, cryptographic verifiability, and zero hallucination.

### Brand Personality & Core Virtues
- **Hyper-Rigorous & Verifiable**: Every statement rendered within the interface is coupled with its official provenance (BOE identifiers, consolidation stamps, and article hashes).
- **Institutional Gravitas meets Terminal Precision**: Visual solemnity derived from archival legal journals paired with the ergonomic minimalism of modern developer platforms (Linear, Vercel, Stripe CLI).
- **Audit-First Utilitarianism**: Aesthetic restraint that avoids decorative embellishment, directing focus toward validity badges, diff comparisons, statutory timelines, and unambiguous status signaling.

### Visual Style
The aesthetic fuses **Corporate/Modern Technical** discipline with **Editorial Typography**:
1. **Pristine Archival Canvas**: Clean, paper-like ivory and mist backgrounds with hairline borders that evoke archival legal registers and high-density IDE panes.
2. **Subtle Monastic Structure**: Explicit containment lines, strict vertical grids, and tabular layouts that communicate structured, immutable data.
3. **High-Contrast Signal Badges**: Dense, high-visibility status chips with sharp monospaced tags that immediately alert autonomous agents and legal engineers to statutory invalidity or regulatory repeal.

## Colors

The color system reconciles judicial authority with developer operations. The hierarchy prioritizes textual readability, official statutory validation, and deterministic alerts.

### Palette Architecture
- **Primary (`#0B1325` - Slate Midnight)**: Anchor tone used for primary headings, dominant structural frames, prominent navigation headers, and core action surfaces. Imparts sovereign authority and institutional permanence.
- **Secondary (`#D97706` - Amber Seal)**: The color of notary seals, archival marginalia, and legislative stamps. Applied to metadata tags, active tab indicators, citation callouts, and key interaction focal points.
- **Tertiary (`#2563EB` - Agent Blue)**: Signifies real-time AI tool activity, protocol endpoints (MCP), WebSocket synchronization, and code-level references.
- **Neutral (`#64748B` - Slate)**: Calibrated for balanced contrast on ivory canvases, supporting secondary copy, statutory dates, table outlines, and breadcrumb dividers.

### Statutory State Semantics
Legal correctness is binary and mission-critical. Three unambiguous states govern legal norms:
1. **Norma Vigente (`#10B981`)**: Law in force, consolidated and active.
2. **Vigencia Agotada / Modificación Pendiente (`#F59E0B`)**: Expired validity, transitory provision in effect, or unconfirmed consolidation.
3. **Norma Derogada (`#EF4444`)**: Repealed or invalidated provision; requires prominent, unmistakable visual alerts across both human and agent interfaces.

### Dark Code Shell
Terminal panels, code viewers, and CLI outputs employ dedicated dark tokens (`code-dark-bg: #0A0E17`, `code-dark-border: #1E293B`) to establish a clean visual separation between statutory text and programmatic execution.

## Typography

The typographic system creates clear boundaries across three distinct operational layers:

1. **Legislative Authority (Newsreader)**: Used for high-level headings, statutory provisions, legal articles, and preamble text. It conveys intellectual gravitas, encourages sustained reading, and evokes traditional legal gazettes.
2. **Interface Clarity (Inter)**: Handles chrome, navigation, dashboard data, form inputs, metadata summaries, and general body copy. Neutral and systematic, it ensures high legibility at dense scales.
3. **Cryptographic & Syntactic Rigor (JetBrains Mono)**: Handles all machine-verifiable data: BOE identifiers (`[BOE-A-2015-10565, bloque a21]`), legal citation tokens, SHA hashes, MCP protocol endpoints, and terminal outputs.

### Typographic Rules & Hierarchy
- **Legal Article Text**: Rendered using `legal-article` (`Newsreader`, 17px, line height 1.75) to maximize sustained legibility in complex statutes, with slightly relaxed letter spacing.
- **Citation Badges & Hashes**: Always uppercase or structured monospace using `citation-stamp`, tracking at `+0.06em` to maintain legibility in compact tags and badges.
- **Microcopy & Metadata**: Never drop below 11px (`0.6875rem`). Legal disclaimers and validation timestamps must remain legible to human auditors.

## Layout & Spacing

The layout is built on a 12-column responsive fluid grid designed for dual consumption: human legal engineers reviewing briefs and developers inspecting agent reasoning trees.

### Grid & Canvas Structure
- **Desktop (>= 1280px)**: 12-column layout, max-width `1440px`, centered with `2rem` outer margins and `1.5rem` gutters.
- **Split Workspace Layout**:
  - *Primary Stage (8 cols)*: Legal text, comparative diffs, article citations, and analytical summaries.
  - *Inspector Rail (4 cols)*: Real-time agent provenance, BOE validation logs, tool-call executions, and metadata badges.
- **Mobile (< 768px)**: Collapses to a single fluid column with `1rem` outer margin, converting the Inspector Rail into an anchored bottom sheet or stacked drawer.

### Spacing Philosophy
Spacing operates on a disciplined 4px/8px base unit. Density is balanced: tight and compact around technical and tabular data (`space-xs` to `space-sm`), yet generous around long-form statutory reading areas (`space-xl` to `space-2xl`) to support focused legal analysis.

## Elevation & Depth

This design system avoids heavy shadows, blur effects, or whimsical floating elevations. Instead, depth is communicated through **structural containment, tonal stratification, and crisp hairline borders**.

### Layering Architecture
1. **Level 0 (Base Canvas)**: `#F8FAFC` (Canvas Ivory). The primary ground on which all cards and document sheets rest.
2. **Level 1 (Card & Article Surfaces)**: `#FFFFFF` (Pristine White) framed by a 1px solid `#E2E8F0` border. This serves as the reading surface for legal texts, search hits, and code snippets.
3. **Level 2 (Terminal Shell & Inspector Panes)**: `#0A0E17` surface with a `#1E293B` perimeter line. Elevated strictly through contrast rather than shadows, isolating technical execution from statutory interpretation.
4. **Level 3 (Modal Overlays & Critical Alerts)**: `#FFFFFF` with an ambient bounding box: `0 4px 20px -2px rgba(11, 19, 37, 0.08), 0 2px 6px -1px rgba(11, 19, 37, 0.04)`. Accompanied by a 20% opacity `#0B1325` backdrop tint to preserve focus on mission-critical confirmation dialogues.

### Structural Hairlines
Separators, table headers, and split panes strictly employ 1px solid lines using `#E2E8F0` in light containers and `#1E293B` in dark terminal environments.

## Shapes

The interface employs a **Soft (`1`)** roundedness profile (base radius `0.25rem` / `4px`), communicating precision, stability, and technical discipline.

### Radius Distribution
- **Controls & Micro-Badges (4px - `rounded`)**: Buttons, text inputs, citation pills, status badges, and inline code markers. These sharper bounds feel deliberate and utilitarian.
- **Cards, Panels & Code Containers (8px - `rounded-lg`)**: Legal document viewports, citation cards, CLI execution blocks, and table containers.
- **Overlays & Dialogs (12px - `rounded-xl`)**: Modal windows, command palettes (`Cmd+K`), and comprehensive legal comparison drawers.
- **Never Pill-Shaped**: Rounded pill designs (`9999px`) are intentionally avoided, preserving the technical and institutional aesthetic of the interface.

## Components

### 1. Buttons
- **Primary Action (Authority)**:
  - Background: `#0B1325` (Slate Midnight), Text: `#FFFFFF`, Border: 1px solid `#0B1325`.
  - Padding: `0.5rem 1rem`, Font: Inter 14px Semibold.
  - Hover: Background `#0F172A`, shadow subtle: `0 1px 2px rgba(0,0,0,0.05)`.
  - Active: Scale `0.99`.
- **Secondary / Technical Action**:
  - Background: `#FFFFFF`, Text: `#0B1325`, Border: 1px solid `#E2E8F0`.
  - Hover: Background `#F8FAFC`, Border: `#CBD5E1`.
- **Terminal Run Action (Agent Hook)**:
  - Background: `#2563EB`, Text: `#FFFFFF`, Font: JetBrains Mono 13px Medium.
  - Prefix with small prompt indicator glyph `>` or terminal icon.

### 2. Citation Chips & Status Badges
- **Statutory Status Badge**:
  - Structure: Monospace `[⚠ NORMA DEROGADA]` or `[✓ NORMA VIGENTE]`.
  - Padding: `0.1875rem 0.5rem`, Border radius: `4px`, Font: JetBrains Mono 11px Bold, tracking `+0.05em`.
  - *Vigente*: Background `#ECFDF5`, Border `#A7F3D0`, Text `#065F46`.
  - *Agotada/Duda*: Background `#FFFBEB`, Border `#FDE68A`, Text `#92400E`.
  - *Derogada*: Background `#FEF2F2`, Border `#FECACA`, Text `#991B1B`.
- **Citation Anchor Pill**:
  - Structure: `[BOE-A-2015-10565 § art. 21]`.
  - Background: `#F1F5F9`, Border: 1px solid `#CBD5E1`, Text: `#0F172A`, Font: JetBrains Mono 11px Medium.
  - Hover: Border color transitions to `#D97706` (Amber Seal) to reveal citation provenance preview on hover.

### 3. Legal Article Cards & Split Panes
- **Article Container**:
  - Background: `#FFFFFF`, Border: 1px solid `#E2E8F0`, Radius: `8px`.
  - Header: Integrated toolbar with official title (`Newsreader` 16px Bold), Article ID pill, and direct BOE deep-link icon.
  - Left Border Indicator: 3px solid indicator matching statutory state (Green for active, Red for repealed).
  - Body Content: Rendered in `legal-article` typography with selectable legal paragraphs and persistent paragraph numbers.

### 4. Interactive Inputs & Search Bars
- **Legal Search & Natural Language Query**:
  - Background: `#FFFFFF`, Border: 1px solid `#CBD5E1`, Radius: `6px`.
  - Focus Ring: 1px solid `#2563EB` with an ambient glow (`0 0 0 3px rgba(37, 99, 235, 0.1)`).
  - Right Adornment: Monospaced shortcut tag `/` or `⌘K` with `#F1F5F9` background and `#64748B` border.

### 5. Checkboxes, Radios & Switches
- **Checkboxes**:
  - 16x16px, 3px border-radius, 1.5px border `#94A3B8`.
  - Checked: Background `#0B1325`, checkmark crisp white.
- **Radio Buttons**:
  - 16x16px circle, 1.5px border `#94A3B8`. Checked: Solid inner pip `#0B1325` with 3px clearance.

### 6. Agent Execution & Terminal Log Pane
- **Agent Output Block**:
  - Background: `#0A0E17`, Border: 1px solid `#1E293B`, Radius: `8px`.
  - Top Bar: Displaying agent status (`claude-code-mcp: active`), active task hash, latency metrics, and copy button.
  - Content: JetBrains Mono 13px, syntax highlighting supporting JSON, Diff, and Markdown statutory quotes.