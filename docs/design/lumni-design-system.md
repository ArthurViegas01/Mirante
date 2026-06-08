# Lumni Design System — Portable Specification

> **Light to your ideas.**
> A single, self-contained spec for re-implementing the Lumni Design System in **Svelte + modern CSS**.
> Build target: clean role-based CSS custom properties, light **and** dark themes, no framework lock-in.

**How to read this doc**
- Sections 1–7 are the *visual specification* (rules, values, states).
- Section 8 is the *drop-in token block* — copy it into a global stylesheet and wire components to the role tokens (`--color-text`, `--space-4`, …), never to raw hex.
- The brand voice is **calm, declarative, grown-up**. Sentence case everywhere; no emoji in product UI; status is color + icon, never 🎉/⚠️.

---

## 1. Brand / Logo lockup

### Anatomy
The lockup is **mark + wordmark**, set on a single baseline.

| Element | Spec |
|---|---|
| **Mark** | Geometric 4-point spark (the "lumen source" — a light burst). Single flat color, never gradient, never with effects. A faint center dot (Glow on light, white on dark) appears only at ≥32px. |
| **Wordmark** | `lumni`, set in **Instrument Serif _italic_**, weight 400, tracking `-0.01em`, all lowercase. Optical size ≈ 0.8× the mark height. |
| **UI / supporting text** | **Figtree** — never set the wordmark in Figtree, and never set body UI in Instrument Serif. |
| **Lockup gap** | Mark-to-wordmark gap = `0.3 × mark height` (e.g. 44px mark → ~13px gap). |

### Clearspace & minimum size
- **Clearspace:** keep a margin equal to the **mark height** clear on all four sides. Nothing — text, edges, other logos — enters that zone.
- **Minimum size:** wordmark lockup no smaller than **120px wide** (≈ mark at 28px). Below that, drop the wordmark and use **mark-only** (minimum **16×16**).

### Color variants
| Variant | Surface | Mark color | Wordmark color |
|---|---|---|---|
| **Beam (default)** | Frost / Mist (light) | Beam `#0F766E` | Ink-950 `#0B1622` |
| **Glow (dark)** | Ink-950 (dark) | Glow `#5EEAD4` | White `#FFFFFF` |
| **Ink (mono)** | Light, when teal is occupied | Ink-950 `#0B1622` | Ink-950 |
| **Reverse** | On Beam `#0F766E` fill | White | White |

### Do / Don't
- ✅ Reproduce in a single flat color. ✅ Use Glow mark only on Ink surfaces. ✅ Keep the full clearspace.
- ❌ No gradient, glow, drop-shadow, or bevel on the mark. ❌ Never set the wordmark in Figtree, Geist Mono, or any non-italic serif. ❌ Don't recolor the wordmark to Beam (wordmark is Ink or white only). ❌ Don't place the lockup on a busy photo without a flat color layer behind it. ❌ Don't stretch, rotate, or re-space the lockup.

---

## 2. Color

Lumni is **cool-toned** — teal brand, navy-grey neutrals, never warm grey. **One brand color on screen at a time** (Beam *or* Tide as accent, never both). Glow is decorative only — never body text on Glow, never Glow as a button fill. Semantic colors are for status only, never brand expression.

### 2.1 Brand ramps (theme-independent constants)

**Beam — teal (brand + interactive)**
| Token | Hex | Note |
|---|---|---|
| `--beam-50` | `#ECFDF5` | tint background |
| `--beam-100` | `#D1FAF0` | tint border |
| `--beam-200` | `#A7F3DD` | |
| `--beam-300` | `#6FE7CB` | dark-theme primary hover |
| `--beam-400` | `#34D3B3` | **dark-theme primary** |
| `--beam-500` | `#14B8A6` | **Tide** — focus ring, accent |
| `--beam-600` | `#0F8E83` | primary hover (light) |
| `--beam-700` | `#0F766E` | **Beam** — brand primary, primary button |
| `--beam-800` | `#115E59` | primary active (light) |
| `--beam-900` | `#134E4A` | |
| `--beam-950` | `#042F2E` | |
| `--glow` | `#5EEAD4` | mint light accent |

**Ink — cool navy neutrals**
| Token | Hex | Role anchor |
|---|---|---|
| `--ink-50` | `#F7F9FA` | dark-theme text |
| `--ink-100` | `#EEF2F4` | subtle divider (light) |
| `--ink-200` | `#DCE3E8` | default border (light) |
| `--ink-300` | `#C0CAD2` | strong border (light) |
| `--ink-400` | `#94A3AE` | placeholder / disabled / dark-theme muted text |
| `--ink-500` | `#6B7A86` | tertiary text, icons |
| `--ink-600` | `#4B5A66` | |
| `--ink-700` | `#334452` | secondary text (light) / dark border-strong |
| `--ink-800` | `#1F2D3A` | dark surface-elevated |
| `--ink-900` | `#0F1B27` | dark surface |
| `--ink-950` | `#0B1622` | primary text (light) / dark background |

**Surfaces (light):** `--frost #FFFFFF` (cards, modals) · `--mist #F4F7F8` (page bg) · `--vapor #FAFBFC` (alt subtle bg).

### 2.2 Semantic status colors

| Role | Light fg | Light bg (subtle) | Dark fg (lightened) |
|---|---|---|---|
| **Success** | `#16A34A` | `#DCFCE7` | `#4ADE80` |
| **Warning** | `#D97706` | `#FEF3C7` | `#FBBF24` |
| **Danger** | `#DC2626` | `#FEE2E2` | `#F87171` |
| **Info** | `#2563EB` | `#DBEAFE` | `#60A5FA` |

On a status pill, text uses the **deeper** shade for legibility on the subtle bg (e.g. On track → text `#15803D` on `#DCFCE7`).

### 2.3 Semantic role tokens (these flip per theme)

| Role | Light | Dark |
|---|---|---|
| `--color-bg` | `#F4F7F8` (Mist) | `#0B1622` (Ink-950) |
| `--color-surface` | `#FFFFFF` (Frost) | `#0F1B27` (Ink-900) |
| `--color-surface-elevated` | `#FFFFFF` + shadow | `#1F2D3A` (Ink-800) |
| `--color-border` | `#DCE3E8` (Ink-200) | `rgba(255,255,255,.08)` |
| `--color-border-strong` | `#C0CAD2` (Ink-300) | `#334452` (Ink-700) |
| `--color-divider` | `#EEF2F4` (Ink-100) | `rgba(255,255,255,.05)` |
| `--color-text` | `#0B1622` (Ink-950) | `#F7F9FA` (Ink-50) |
| `--color-text-secondary` | `#334452` (Ink-700) | `rgba(255,255,255,.72)` |
| `--color-text-muted` | `#6B7A86` (Ink-500) | `#94A3AE` (Ink-400) |
| `--color-text-disabled` | `#94A3AE` (Ink-400) | `rgba(255,255,255,.4)` |
| `--color-primary` | `#0F766E` (Beam) | `#34D3B3` (Beam-400) |
| `--color-primary-hover` | `#115E59` (Beam-800) | `#6FE7CB` (Beam-300) |
| `--color-primary-text` | `#FFFFFF` | `#0B1622` |
| `--color-accent` | `#14B8A6` (Tide) | `#5EEAD4` (Glow) |
| `--color-link` | `#0F766E` | `#5EEAD4` |

> **Why the brand lightens in dark:** Beam-700 on Ink-950 fails contrast. On dark surfaces, primary shifts up the ramp to Beam-400 / Glow, and primary buttons take **Ink-950 text** (not white).

### 2.4 Chart / data palette
Teal-forward categorical ramp — disciplined and cool, with three distinct hues for multi-series. Order = series priority.

| Token | Hex | Use |
|---|---|---|
| `--chart-1` | `#0F766E` | primary series (Beam) |
| `--chart-2` | `#14B8A6` | Tide |
| `--chart-3` | `#60A5FA` | sky |
| `--chart-4` | `#F59E0B` | amber |
| `--chart-5` | `#8B5CF6` | violet |
| `--chart-6` | `#5EEAD4` | Glow (highlight / "live" point) |
| `--chart-grid` | `#EEF2F4` (light) / `rgba(255,255,255,.06)` (dark) | gridlines |
| `--chart-baseline` | `#DCE3E8` @ 50% (light) | filled baseline area |

Line series: `2.2px` stroke, rounded joins. Area fill: series color at **6% opacity**. Live/last point: filled dot `r=4` + halo `r=8` at 20% opacity.

### 2.5 Contrast & focus

**Key pairs (≈ WCAG 2.1 ratios, approximate):**
| Foreground | Background | Ratio | Verdict |
|---|---|---|---|
| Ink-950 `#0B1622` | Frost `#FFFFFF` | ≈ 18.8 : 1 | AAA |
| Ink-700 `#334452` | Frost | ≈ 9.1 : 1 | AAA |
| Ink-500 `#6B7A86` | Frost | ≈ 4.4 : 1 | AA normal (borderline) — use for tertiary/large, not long body |
| White | Beam-700 `#0F766E` | ≈ 5.5 : 1 | AA — primary button OK |
| White | Tide `#14B8A6` | ≈ 2.3 : 1 | **Fail** — never white text on Tide; Tide is for rings/accents, not text bg |
| Ink-50 `#F7F9FA` | Ink-950 (dark bg) | ≈ 16 : 1 | AAA |
| Glow `#5EEAD4` | Ink-950 | ≈ 11 : 1 | AAA — dark links/accents |

**Focus ring (always visible on keyboard focus, never suppressed):**
```
box-shadow: 0 0 0 3px rgba(20, 184, 166, 0.24);   /* 3px Tide ring */
```
On dark surfaces, raise alpha to ~0.4 for the same Tide hue.

---

## 3. Typography

### Families
| Token | Family | Role |
|---|---|---|
| `--font-sans` | **Figtree** | The workhorse — all UI from 12px labels to 84px display headlines. |
| `--font-serif` | **Instrument Serif** _italic_ | Display/editorial **only** — wordmark, pull quotes, a single italic word in a hero, slide section breaks. **Never at body size.** |
| `--font-mono` | **Geist Mono** | Code, numerals in data/tables, technical UI (paths, IDs, timestamps), and the eyebrow caps label. Use tabular figures (`font-feature-settings: "tnum"`) wherever numbers are compared. |

Fallback stacks: sans → `ui-sans-serif, system-ui, -apple-system, sans-serif`; serif → `ui-serif, Georgia, serif`; mono → `ui-monospace, "SF Mono", Menlo, monospace`.

### When serif vs sans
- **Sans (Figtree)** is the default for *everything* functional. If in doubt, use sans.
- **Serif (Instrument Serif italic)** is a deliberate editorial accent — one moment per view, at large size. A pull quote, a hero emphasis word, a section break. It signals warmth/craft, not hierarchy.

### Type scale
Base = 16px (`1rem`). Display weight is **medium (500)**, not bold — Lumni is precise, not loud. 700 is reserved; 800/900 exist in the font but are off-system.

| Level | Size | Line-height | Weight | Tracking | Font |
|---|---|---|---|---|---|
| **Display** | 84px / 5.25rem | 1.02 | 500 | -0.03em | Figtree |
| **Display-serif** | 64–84px | 1.05 | 400 italic | -0.01em | Instrument Serif |
| **H1** | 48px / 3rem | 1.1 | 500 | -0.02em | Figtree |
| **H2** | 38px / 2.375rem | 1.1 | 500 | -0.02em | Figtree |
| **H3** | 30px / 1.875rem | 1.25 | 500 | -0.01em | Figtree |
| **H4** | 24px / 1.5rem | 1.25 | 500 | -0.01em | Figtree |
| **Body-lg / lead** | 17px / 1.0625rem | 1.65 | 400 | 0 | Figtree |
| **Body** | 15px / 0.9375rem | 1.65 | 400 | 0 | Figtree |
| **Small** | 13px / 0.8125rem | 1.5 | 400 | 0 | Figtree |
| **Caption / eyebrow** | 12px / 0.75rem | 1.4 | 500 | +0.12em, UPPERCASE | Figtree or Mono |
| **Mono / data** | 11–14px | 1.5 | 400–500 | +0.04–0.08em | Geist Mono |

Tracking is **tight on headlines** (`-0.02em`→`-0.03em` at display), **normal on body**. Line-height generous (1.65 body), never cramped.

---

## 4. Spacing, radius, border, elevation

### Spacing — 4px base grid (everything snaps to it)
| Token | px | rem |
|---|---|---|
| `--space-1` | 4 | 0.25 |
| `--space-2` | 8 | 0.5 |
| `--space-3` | 12 | 0.75 |
| `--space-4` | 16 | 1 |
| `--space-5` | 20 | 1.25 |
| `--space-6` | 24 | 1.5 |
| `--space-8` | 32 | 2 |
| `--space-10` | 40 | 2.5 |
| `--space-12` | 48 | 3 |
| `--space-16` | 64 | 4 |
| `--space-20` | 80 | 5 |
| `--space-24` | 96 | 6 |
| `--space-32` | 128 | 8 |

**Rhythm:** section spacing 96px desktop / 64px mobile. Card padding 24px default, 32px feature. Prose max-width 680px; content max 1200px (marketing) / 1440px (product).

### Radius
| Token | px | Use |
|---|---|---|
| `--radius-xs` | 4 | min on any interactive element |
| `--radius-sm` | 6 | small controls |
| `--radius-md` | 8 | **default** — inputs, buttons |
| `--radius-lg` | 12 | cards |
| `--radius-xl` | 16 | feature cards |
| `--radius-2xl` | 24 | hero surfaces, modals |
| `--radius-full` | 9999 | **pills only** (Tag, Badge) |

Never < 4px on interactive elements. No pill buttons except Tag/Badge.

### Border widths
| Token | px | Use |
|---|---|---|
| `--border-width-1` | 1 | default border / divider |
| `--border-width-2` | 2 | avatar rings, active tab underline |
| `--border-width-focus` | 3 | focus ring |

Always cool grey (`--color-border`), never black. No dashed/double/colored borders except focus (3px Tide) and semantic states. On hover, borders deepen to `--color-border-strong`.

### Elevation — 5 tiers, cool-tinted (Ink-950 alpha, **never** rgba(0,0,0))
| Token | Value | Use |
|---|---|---|
| `--shadow-xs` | `0 1px 2px 0 rgba(15,27,39,.04)` | input, checkbox |
| `--shadow-sm` | `0 1px 3px 0 rgba(15,27,39,.06), 0 1px 2px -1px rgba(15,27,39,.04)` | **card resting (default)** |
| `--shadow-md` | `0 4px 8px -2px rgba(15,27,39,.06), 0 2px 4px -2px rgba(15,27,39,.04)` | card hover, dropdown |
| `--shadow-lg` | `0 12px 24px -6px rgba(15,27,39,.08), 0 4px 8px -4px rgba(15,27,39,.04)` | modal, popover |
| `--shadow-xl` | `0 24px 48px -12px rgba(15,27,39,.12)` | hero card (rare) |
| `--shadow-inner` | `inset 0 1px 0 0 rgba(255,255,255,.6)` | special (inputs in dark) |
| `--shadow-focus` | `0 0 0 3px rgba(20,184,166,.24)` | keyboard focus ring |

Default to `--shadow-sm`; escalate only when hierarchy demands. Never stack heavy shadows. In **dark mode**, surfaces separate by *background step* (Ink-900 → Ink-800), not drop-shadow.

---

## 5. Motion

Confident, decelerating, never bouncy. **No animation longer than 400ms.** Fades + slides of **8–16px max**. Modals fade + scale `0.96 → 1.0`. Respect `prefers-reduced-motion` by clamping all durations to 0ms.

| Token | Value | Use |
|---|---|---|
| `--ease-out` | `cubic-bezier(0.22, 1, 0.36, 1)` | **default** — almost everything (entrances, hovers) |
| `--ease-in-out` | `cubic-bezier(0.65, 0, 0.35, 1)` | symmetric moves (toggles, reorders) |
| `--ease-spring` | `cubic-bezier(0.34, 1.56, 0.64, 1)` | emphasis only — badge appears, send confirms. **Never on routine UI.** |
| `--dur-instant` | 80ms | cursor-driven hover, press (must feel immediate) |
| `--dur-fast` | 160ms | **default** UI transition |
| `--dur-base` | 240ms | entering panels |
| `--dur-slow` | 400ms | page transitions (max) |

- **Hover (button):** background deepens one ramp step (color, not opacity), `--dur-instant`.
- **Hover (card):** border → strong, shadow `sm → md`, optional `translateY(-2px)` over `--dur-fast`.
- **Press:** scale `0.98` over `--dur-instant`, no color change.
- The spring never overshoots past 1.0 — Lumni does not boing.

---

## 6. Iconography

- **Set:** Lucide-style stroke icons (⚠️ substitution — no brand-owned set yet; match Lucide's geometry when commissioning custom).
- **Style:** geometric, **stroke only, never filled**. Filled is reserved for the brand mark and *active* tab state (filled vs. outlined).
- **Grid:** 24×24. **Stroke width 1.75–2px**, rounded caps + joins, ~4px corner radius on strokes.
- **Sizes:** 24px in headers / feature blocks · **20px default** in body UI · 16px in dense lists/tables.
- **Color:** default `--color-text-muted` (functional, not decorative). Brand/semantic color only for feature highlights and status.
- **No emoji as iconography**, ever. Typographic Unicode (→ ↗ · —) *is* allowed in copy; ↗ is the standard external-link affordance.

---

## 7. Components

All measurements from the live Lumni Console kit. States are listed as **default / hover / active / focus / disabled** where applicable.

### 7.1 Navigation sidebar
- **Container:** width `232px`, full height, bg **Ink-950**, right border `1px rgba(255,255,255,.05)`, padding `18px 14px`, item gap `4px`.
- **Brand row:** mark `22px` (Glow) + wordmark 16px/500 white + mono `CONSOLE` tag at `rgba(255,255,255,.4)`.
- **Search field:** bg `rgba(255,255,255,.06)`, border `1px rgba(255,255,255,.08)`, radius `8px`, padding `7px 10px 7px 32px`; leading search glyph + trailing `⌘K` kbd chip.
- **Section label:** mono, 10.5px, `+0.12em`, uppercase, `rgba(255,255,255,.4)`, padding `14px 8px 6px`.
- **Nav item:** padding `7px 9px`, radius `7px`, icon `16px` + label 13.5px, optional trailing mono count.
  - *default* text `rgba(255,255,255,.72)`, icon `rgba(255,255,255,.5)`
  - *hover* bg `rgba(255,255,255,.04)`, text → white (120ms ease-out)
  - *active* bg `rgba(94,234,212,.10)`, text + icon → **Glow** `#5EEAD4`
- **Footer (user):** top border `1px rgba(255,255,255,.06)`; 26px avatar (Beam→Glow gradient) + name 12.5px/500 + org in mono + chevron.

### 7.2 Stat / KPI card
- **Container:** bg `--color-surface`, border `1px --color-border`, radius `--radius-lg` (12px), padding `18px 20px`.
- **Label:** mono, 11px, `+0.08em`, uppercase, `--color-text-muted`.
- **Number:** mono, **28px / 500**, tabular figures, tracking `-0.015em`, `--color-text`, top margin 8px, line-height 1.
- **Delta:** mono, 11.5px, top margin 8px — `up` green `#15803D` (↑), `down` red `#B91C1C` (↓), `flat` muted (→).
- *hover (if interactive):* border → strong, shadow `sm → md`.

### 7.3 Data table
- **Wrapper:** bg `--color-surface`, border `1px --color-border`, radius `12px`, `overflow:hidden`.
- **Head th:** bg `--vapor`, mono 11px `+0.08em` uppercase, `--color-text-muted`, weight 500, padding `11px 18px`, bottom border `1px --color-border`.
- **Body td:** padding `14px 18px`, 13.5px, bottom border `1px --color-divider`; last row no border. Primary cell `.name` weight 600 + `.meta` sub-line in muted 12px. Numerals → mono tabular, right-aligned.
- **Row states:** *default* surface · *hover* bg `--vapor`, `cursor:pointer` (120ms) · *selected* bg `--beam-50` (light) / Ink-800 (dark).

### 7.4 Status badges (pills)
Radius `--radius-full`, padding `2px 9px`, 11px / weight 600, leading 6px dot.

| Status | Text | Bg | Dot |
|---|---|---|---|
| **On track** | `#15803D` | `#DCFCE7` | `#16A34A` |
| **At risk** | `#92400E` | `#FEF3C7` | `#D97706` |
| **Planning** | `#1E3A8A` | `#DBEAFE` | `#2563EB` |
| **Blocked** | `#991B1B` | `#FEE2E2` | `#DC2626` |

Text + dot are the status color; bg is its subtle tint. No emoji — the dot + label carry meaning.

### 7.5 Progress bar
- **Track:** height `5px`, bg `--ink-100`, radius `--radius-full`, `overflow:hidden`.
- **Fill:** full-height span, width = %, bg `--beam-500` (Tide) by default.
- **Semantic fill:** at-risk → `#D97706`, blocked → `#DC2626` (mirror the row's status).

### 7.6 Sparkline / chart
- **Frame:** inside a `--panel` (surface + border + 12px radius); body padding 20px; chart height ~188px, `preserveAspectRatio:none` SVG.
- **Gridlines:** `--chart-grid`, 1px.
- **Primary line:** `--chart-1` (Beam), `2.2px`, rounded joins; **area fill** same hue @ **6%**.
- **Baseline series:** filled area `--chart-baseline` (Ink-200 @ 50%).
- **Live point:** filled dot `r=4` + halo `r=8` @ 20%.
- **Axis labels:** mono, 10.5px, `--color-text-muted`.
- **Legend:** inline dots in panel head; active series full-color, secondary muted.

### 7.7 Activity feed
- **Item:** padding `12px 20px`, bottom divider `1px --color-divider` (last none), 13px / line-height 1.45, leading **8px status dot** (semantic color, top margin 6px).
- **Actor name:** weight 600, `--color-text`.
- **Timestamp:** mono, 11px, `--color-text-muted`, own line.
- Panel head pattern: title 14px/600 + mono sub-label + right-aligned `View all ↗` link in brand color.

### 7.8 Buttons
Radius `--radius-md` (8px), Figtree weight **600**, tracking `-0.005em`, no trailing period. Sizes: **sm** `8px 14px / 13px` · **md** `10px 18px / 14px` · **lg** `14px 24px / 15px`.

| Variant | Default | Hover | Active | Focus | Disabled |
|---|---|---|---|---|---|
| **Primary** | bg Beam `#0F766E`, white text | bg Beam-800 `#115E59` | bg Beam-900 `#134E4A` | + 3px Tide ring | opacity .5, `not-allowed` |
| **Secondary** | surface bg, `1px --color-border`, text `--color-text` | border → strong, bg `--mist` | bg `--ink-100` | + 3px Tide ring | opacity .5 |
| **Ghost** | transparent, text `--color-text-secondary` | bg `--mist`, text `--color-text` | bg `--ink-100` | + 3px Tide ring | opacity .5 |

Hover is a **color step, never opacity**. Press adds scale `0.98` @ 80ms. Dark theme: primary bg → Beam-400 with **Ink-950 text**.

### 7.9 Input / search
- **Text input:** Figtree 13px, padding `8px 10px`, bg `--color-surface`, border `1px --color-border`, radius `--radius-md`, text `--color-text`.
  - *hover* border → `--color-border-strong`
  - *focus* border `--color-primary` + **3px Tide ring** (`--shadow-focus`)
  - *filled/active* border `--color-border-strong`
  - *disabled* opacity .5, `not-allowed`
  - *error* border `--color-danger` + danger-tinted ring
- **Label:** 11.5px / weight 500, `--color-text-secondary`, 4px above field.
- **Select:** same metrics + trailing chevron.
- **Checkbox:** 16px box, radius `--radius-xs`; checked = Beam fill + white check.
- **Switch:** track `30×18px` radius-full; on = Beam fill, 14px white knob inset 2px; off = `--ink-300` track.
- **Search (in dark sidebar):** translucent `rgba(255,255,255,.06)` bg, leading glyph, trailing `⌘K` chip.

---

## 8. Tokens as code

Drop this into a global stylesheet (e.g. `src/app.css`). Raw ramps are theme-independent; **role tokens** flip between `:root` (light) and `[data-theme="dark"]`. Wire components to role tokens only.

```css
/* ============================================================
   LUMNI DESIGN SYSTEM — TOKENS
   Light = :root  ·  Dark = [data-theme="dark"]
   Load fonts first: Figtree (sans), Instrument Serif (serif), Geist Mono (mono)
   ============================================================ */

:root {
  /* ---------- BRAND RAMPS (theme-independent) ---------- */
  --beam-50:  #ECFDF5;
  --beam-100: #D1FAF0;
  --beam-200: #A7F3DD;
  --beam-300: #6FE7CB;
  --beam-400: #34D3B3;
  --beam-500: #14B8A6;   /* Tide */
  --beam-600: #0F8E83;
  --beam-700: #0F766E;   /* Beam */
  --beam-800: #115E59;
  --beam-900: #134E4A;
  --beam-950: #042F2E;
  --glow:     #5EEAD4;

  --ink-50:  #F7F9FA;
  --ink-100: #EEF2F4;
  --ink-200: #DCE3E8;
  --ink-300: #C0CAD2;
  --ink-400: #94A3AE;
  --ink-500: #6B7A86;
  --ink-600: #4B5A66;
  --ink-700: #334452;
  --ink-800: #1F2D3A;
  --ink-900: #0F1B27;
  --ink-950: #0B1622;

  --frost: #FFFFFF;
  --mist:  #F4F7F8;
  --vapor: #FAFBFC;

  /* ---------- SEMANTIC STATUS (light) ---------- */
  --color-success:    #16A34A;
  --color-success-bg: #DCFCE7;
  --color-success-text: #15803D;
  --color-warning:    #D97706;
  --color-warning-bg: #FEF3C7;
  --color-warning-text: #92400E;
  --color-danger:     #DC2626;
  --color-danger-bg:  #FEE2E2;
  --color-danger-text: #991B1B;
  --color-info:       #2563EB;
  --color-info-bg:    #DBEAFE;
  --color-info-text:  #1E3A8A;

  /* ---------- CHART PALETTE ---------- */
  --chart-1: #0F766E;
  --chart-2: #14B8A6;
  --chart-3: #60A5FA;
  --chart-4: #F59E0B;
  --chart-5: #8B5CF6;
  --chart-6: #5EEAD4;
  --chart-grid:     #EEF2F4;
  --chart-baseline: rgba(220, 227, 232, 0.5);

  /* ---------- ROLE TOKENS (light) ---------- */
  --color-bg:               var(--mist);
  --color-surface:          var(--frost);
  --color-surface-elevated: var(--frost);
  --color-border:           var(--ink-200);
  --color-border-strong:    var(--ink-300);
  --color-divider:          var(--ink-100);
  --color-text:             var(--ink-950);
  --color-text-secondary:   var(--ink-700);
  --color-text-muted:       var(--ink-500);
  --color-text-disabled:    var(--ink-400);
  --color-primary:          var(--beam-700);
  --color-primary-hover:    var(--beam-800);
  --color-primary-active:   var(--beam-900);
  --color-primary-text:     #FFFFFF;
  --color-accent:           var(--beam-500);
  --color-link:             var(--beam-700);
  --color-link-hover:       var(--beam-800);
  --color-focus-ring:       rgba(20, 184, 166, 0.24);

  /* ---------- TYPOGRAPHY ---------- */
  --font-sans:  "Figtree", ui-sans-serif, system-ui, -apple-system, sans-serif;
  --font-serif: "Instrument Serif", ui-serif, Georgia, serif;
  --font-mono:  "Geist Mono", ui-monospace, "SF Mono", Menlo, monospace;

  --text-display: 5.25rem;   /* 84px */
  --text-4xl: 3rem;          /* 48px  H1 */
  --text-3xl: 2.375rem;      /* 38px  H2 */
  --text-2xl: 1.875rem;      /* 30px  H3 */
  --text-xl:  1.5rem;        /* 24px  H4 */
  --text-lg:  1.25rem;       /* 20px */
  --text-md:  1.0625rem;     /* 17px  lead */
  --text-base: 0.9375rem;    /* 15px  body */
  --text-sm:  0.8125rem;     /* 13px  small */
  --text-xs:  0.75rem;       /* 12px  caption/eyebrow */

  --leading-tight:   1.1;
  --leading-snug:    1.25;
  --leading-normal:  1.5;
  --leading-relaxed: 1.65;
  --leading-display: 1.02;

  --tracking-display: -0.03em;
  --tracking-tight:   -0.02em;
  --tracking-snug:    -0.01em;
  --tracking-normal:  0;
  --tracking-wide:    0.04em;
  --tracking-eyebrow: 0.12em;

  --weight-regular:  400;
  --weight-medium:   500;
  --weight-semibold: 600;
  --weight-bold:     700;

  /* ---------- SPACING (4px grid) ---------- */
  --space-1:  0.25rem;   /* 4   */
  --space-2:  0.5rem;    /* 8   */
  --space-3:  0.75rem;   /* 12  */
  --space-4:  1rem;      /* 16  */
  --space-5:  1.25rem;   /* 20  */
  --space-6:  1.5rem;    /* 24  */
  --space-8:  2rem;      /* 32  */
  --space-10: 2.5rem;    /* 40  */
  --space-12: 3rem;      /* 48  */
  --space-16: 4rem;      /* 64  */
  --space-20: 5rem;      /* 80  */
  --space-24: 6rem;      /* 96  */
  --space-32: 8rem;      /* 128 */

  /* ---------- RADIUS ---------- */
  --radius-xs:   4px;
  --radius-sm:   6px;
  --radius-md:   8px;
  --radius-lg:   12px;
  --radius-xl:   16px;
  --radius-2xl:  24px;
  --radius-full: 9999px;

  /* ---------- BORDER WIDTH ---------- */
  --border-width-1:     1px;
  --border-width-2:     2px;
  --border-width-focus: 3px;

  /* ---------- ELEVATION (cool-tinted) ---------- */
  --shadow-xs: 0 1px 2px 0 rgba(15,27,39,.04);
  --shadow-sm: 0 1px 3px 0 rgba(15,27,39,.06), 0 1px 2px -1px rgba(15,27,39,.04);
  --shadow-md: 0 4px 8px -2px rgba(15,27,39,.06), 0 2px 4px -2px rgba(15,27,39,.04);
  --shadow-lg: 0 12px 24px -6px rgba(15,27,39,.08), 0 4px 8px -4px rgba(15,27,39,.04);
  --shadow-xl: 0 24px 48px -12px rgba(15,27,39,.12);
  --shadow-inner: inset 0 1px 0 0 rgba(255,255,255,.6);
  --shadow-focus: 0 0 0 3px var(--color-focus-ring);

  /* ---------- MOTION ---------- */
  --ease-out:    cubic-bezier(0.22, 1, 0.36, 1);
  --ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
  --dur-instant: 80ms;
  --dur-fast:    160ms;
  --dur-base:    240ms;
  --dur-slow:    400ms;

  /* ---------- LAYOUT ---------- */
  --max-content: 1200px;
  --max-canvas:  1440px;
  --max-prose:   680px;
  --gutter:      24px;
}

/* ============================================================
   DARK THEME — role tokens flip; ramps/scale stay constant
   ============================================================ */
[data-theme="dark"] {
  --color-bg:               var(--ink-950);
  --color-surface:          var(--ink-900);
  --color-surface-elevated: var(--ink-800);
  --color-border:           rgba(255,255,255,0.08);
  --color-border-strong:    var(--ink-700);
  --color-divider:          rgba(255,255,255,0.05);
  --color-text:             var(--ink-50);
  --color-text-secondary:   rgba(255,255,255,0.72);
  --color-text-muted:       var(--ink-400);
  --color-text-disabled:    rgba(255,255,255,0.40);

  --color-primary:          var(--beam-400);
  --color-primary-hover:    var(--beam-300);
  --color-primary-active:   var(--glow);
  --color-primary-text:     var(--ink-950);   /* dark text on light teal */
  --color-accent:           var(--glow);
  --color-link:             var(--glow);
  --color-link-hover:       var(--beam-300);
  --color-focus-ring:       rgba(20, 184, 166, 0.40);

  /* status — lightened for legibility on dark */
  --color-success:      #4ADE80;
  --color-success-bg:   rgba(74,222,128,0.14);
  --color-success-text: #4ADE80;
  --color-warning:      #FBBF24;
  --color-warning-bg:   rgba(251,191,36,0.14);
  --color-warning-text: #FBBF24;
  --color-danger:       #F87171;
  --color-danger-bg:    rgba(248,113,113,0.14);
  --color-danger-text:  #F87171;
  --color-info:         #60A5FA;
  --color-info-bg:      rgba(96,165,250,0.14);
  --color-info-text:    #60A5FA;

  /* chart grid/baseline for dark */
  --chart-grid:     rgba(255,255,255,0.06);
  --chart-baseline: rgba(255,255,255,0.08);

  /* elevation: prefer background-step separation; soften shadows */
  --shadow-sm: 0 1px 3px 0 rgba(0,0,0,.4);
  --shadow-md: 0 4px 8px -2px rgba(0,0,0,.45);
  --shadow-lg: 0 12px 24px -6px rgba(0,0,0,.5);
}

/* Accessibility: clamp motion */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0ms !important;
    transition-duration: 0ms !important;
  }
}
```

### Implementation notes for the Svelte team
- **Theme switch:** set `data-theme="dark"` on `<html>` (or a root wrapper). All role tokens cascade automatically — no per-component dark variants needed.
- **Never hard-code hex in components.** Reference `--color-*`, `--space-*`, `--radius-*`, `--shadow-*`, `--font-*`, `--ease-*`. The raw ramps exist so role tokens can be *re-pointed*, not consumed directly.
- **Fonts:** load Figtree (self-host), Instrument Serif + Geist Mono (Google Fonts or self-host). Apply `font-feature-settings: "tnum"` on any element comparing numerals (tables, KPIs, charts).
- **Focus is non-negotiable:** keep the 3px Tide ring visible on `:focus-visible`. Don't suppress outlines for aesthetics.
- **Density:** body UI 13–15px, dense tables/lists 13.5px, KPI numbers 28px mono. The system reads *quiet and precise* — when unsure, choose more whitespace and less color.
