---
name: Technical Observability System
colors:
  surface: '#051424'
  surface-dim: '#051424'
  surface-bright: '#2c3a4c'
  surface-container-lowest: '#010f1f'
  surface-container-low: '#0d1c2d'
  surface-container: '#122131'
  surface-container-high: '#1c2b3c'
  surface-container-highest: '#273647'
  on-surface: '#d4e4fa'
  on-surface-variant: '#bdc8d1'
  inverse-surface: '#d4e4fa'
  inverse-on-surface: '#233143'
  outline: '#87929a'
  outline-variant: '#3e484f'
  surface-tint: '#7bd0ff'
  primary: '#8ed5ff'
  on-primary: '#00354a'
  primary-container: '#38bdf8'
  on-primary-container: '#004965'
  inverse-primary: '#00668a'
  secondary: '#bcc7de'
  on-secondary: '#263143'
  secondary-container: '#3e495d'
  on-secondary-container: '#aeb9d0'
  tertiary: '#ffc176'
  on-tertiary: '#472a00'
  tertiary-container: '#f1a02b'
  on-tertiary-container: '#613b00'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#c4e7ff'
  primary-fixed-dim: '#7bd0ff'
  on-primary-fixed: '#001e2c'
  on-primary-fixed-variant: '#004c69'
  secondary-fixed: '#d8e3fb'
  secondary-fixed-dim: '#bcc7de'
  on-secondary-fixed: '#111c2d'
  on-secondary-fixed-variant: '#3c475a'
  tertiary-fixed: '#ffddb8'
  tertiary-fixed-dim: '#ffb960'
  on-tertiary-fixed: '#2a1700'
  on-tertiary-fixed-variant: '#653e00'
  background: '#051424'
  on-background: '#d4e4fa'
  surface-variant: '#273647'
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '700'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.3'
    letterSpacing: -0.01em
  title-sm:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '600'
    lineHeight: '1.4'
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.5'
  body-sm:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: '400'
    lineHeight: '1.5'
  data-lg:
    fontFamily: JetBrains Mono
    fontSize: 18px
    fontWeight: '500'
    lineHeight: '1.2'
  data-md:
    fontFamily: JetBrains Mono
    fontSize: 14px
    fontWeight: '500'
    lineHeight: '1.2'
  label-caps:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: '700'
    lineHeight: '1'
    letterSpacing: 0.05em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  unit: 4px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 48px
  gutter: 16px
  margin-mobile: 16px
  margin-desktop: 32px
---

## Brand & Style

The design system is engineered for high-stakes monitoring environments where cognitive load must be minimized and critical information must be prioritized. The brand personality is **precise, authoritative, and vigilant**. It avoids decorative flourishes in favor of utility and data density, targeting DevOps engineers and SREs who require immediate clarity during system incidents.

The visual style is a fusion of **Corporate Modern and Technical Minimalism**. It utilizes a deep-space background palette to reduce eye strain during long-term monitoring, punctuated by high-chroma status indicators that leverage the psychological associations of color to signal system health. The interface feels like a sophisticated instrument panel: intentional, structured, and uncompromisingly functional.

## Colors

The palette is anchored by a "Deep Charcoal" base to provide a non-distracting canvas for data visualization. 

- **Base Background (#0F172A):** Used for the primary application shell and background.
- **Surface Layer (#1E293B):** Used for cards, panels, and distinct UI sections to create subtle depth.
- **Primary Accent (#38BDF8):** A bright sky blue used for interactive elements and focused states.
- **Status Colors:** These are the most critical colors in the system. They are calibrated for high contrast against the dark background to ensure that "Warning" and "Critical" states are impossible to miss. Success states are vibrant but secondary to alerts.

## Typography

This design system uses a dual-font approach to balance readability with technical precision.

- **Inter:** Used for all UI chrome, navigation, and descriptive text. Its high x-height ensures legibility at small sizes.
- **JetBrains Mono:** Employed specifically for metrics, timestamps, logs, and any numerical data. The monospaced nature allows for easier visual scanning of vertically aligned numbers in dashboards.

**Hierarchy Rules:**
- Use `label-caps` for section headers and table column titles to create clear separation.
- `data-md` is the default for metrics within status cards.
- Mobile scaling: On devices smaller than 768px, `display-lg` should scale down to 24px to maintain screen real estate.

## Layout & Spacing

The layout follows a **Strict Fluid Grid** model based on a 4px baseline. 

- **Grid:** Use a 12-column grid for desktop views. Elements should primarily span 3, 4, or 6 columns to maintain a balanced dashboard look.
- **Density:** This is a high-density system. Vertical spacing between related data points should be `sm` (8px), while spacing between unrelated modules should be `lg` (24px).
- **Responsive Behavior:** 
  - **Desktop (>1024px):** Side navigation is persistent. Margins are `xl`.
  - **Tablet (768px - 1023px):** Side navigation collapses to icons. Margins are `lg`.
  - **Mobile (<767px):** Content reflows into a single column. Horizontal padding is reduced to `md`.

## Elevation & Depth

To maintain a professional, data-first aesthetic, this design system avoids heavy drop shadows. Instead, it utilizes **Tonal Layering and Low-Contrast Outlines**.

- **Level 0 (Background):** `#0F172A` - The base of the application.
- **Level 1 (Surface):** `#1E293B` - Used for widgets and cards. These should have a subtle 1px border of `#334155` to define their edges.
- **Level 2 (Overlay):** `#1E293B` - Used for modals and dropdowns. These are distinguished by a slightly lighter 1px border (`#475569`) and a soft, deep ambient shadow (20% opacity, 12px blur) to suggest height.
- **Active State:** Focused inputs or active tabs use a glow effect derived from the primary color but with very low opacity (15%) to avoid visual noise.

## Shapes

The shape language is **Soft and Disciplined**. Elements use a 4px (0.25rem) base radius to appear modern without feeling overly casual or "bubbly."

- **Standard Elements:** Buttons, input fields, and small cards use the `rounded-md` (4px).
- **Large Containers:** Main dashboard widgets use `rounded-lg` (8px).
- **Status Badges:** Small pills or status indicators may use a fully rounded (pill) shape to differentiate them from interactive buttons.

## Components

**Buttons:**
- **Primary:** Solid `#38BDF8` with dark text. 
- **Secondary:** Transparent with `#334155` border and white text.
- **Ghost:** No border, `#94A3B8` text, becoming white on hover.

**Status Chips:**
- High-contrast background with dark text. For example, a "Critical" chip uses the `#EF4444` background. These must always include an icon (e.g., an exclamation mark) to ensure accessibility for colorblind users.

**Cards (Widgets):**
- Must include a header with `label-caps` text and a 1px bottom border. Internal padding should be a consistent `md` (16px).

**Input Fields:**
- Background should be darker than the surface (`#0F172A`).
- Border becomes `#38BDF8` on focus.
- Error states change the border to `#EF4444`.

**Lists & Tables:**
- Use zebra-striping with a very subtle variance in background color for high-density tables.
- Rows should have a hover state that highlights the entire line in `#334155`.

**Specialized Components:**
- **Sparklines:** Compact, monochromatic trend lines within cards.
- **Uptime Bar:** A horizontal bar segmented by time, using the status colors to show history.