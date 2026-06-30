## Context

The dashboard has accumulated several visual inconsistencies as new modules (channels, policies, service archive) were added:

1. `devicons-react` SVGs ignore CSS `className` sizing and render at intrinsic pixel size, making technology icons look small inside the `bg-surface-high` frame that wraps them. The component already supports `sm | md | lg` sizes via Tailwind classes, but the underlying SVG does not honor them.
2. Timestamps render in the sans typeface on most surfaces even though the dashboard already imports JetBrains Mono and applies it in table cells.
3. The notification channels list shows the type as plain text (`telegram`, `email`, `sms`, `webhook`, `pagerduty`). The channel form already uses a `Select` but offers no visual signal.
4. The shared `AppShell` renders a sticky top header above `main` containing an eyebrow label, a tagline, and a duplicate "Create service" CTA. The sidebar already carries the "Create service" CTA through per-page headers.
5. The `ServiceForm` renders a read-only Lifecycle block that mirrors information already on the service detail summary card and the home service health matrix.

## Goals / Non-Goals

**Goals:**
- Make technology icons fill their frame at every supported size.
- Establish mono font as the default for every rendered timestamp on dashboard surfaces.
- Add a small, recognizable icon for each notification channel type on the channels list.
- Remove the sticky top header from `AppShell`.
- Remove the read-only Lifecycle field from the service form.

**Non-Goals:**
- New product behavior (no new buttons, fields, or flows).
- Changes to backend data models or APIs.
- Replacing `devicons-react` with a different icon library.
- Reorganizing sidebar items.

## Decisions

### Technology icon sizing
- Pass an explicit numeric `size` prop on `devicons-react` icons matching the frame's pixel size (`sm=36`, `md=44`, `lg=56`).
- Keep the existing `bg-surface-high` frame and `brightness-0 invert` filter; only the inner icon changes.
- Rationale: minimal change, library-native. Alternative considered was wrapping each icon in a `<svg>` and overriding `viewBox`, but that fights the library's exported component shape.

### Mono font for timestamps
- Apply `font-mono text-xs` (or appropriate size) to every `formatDateTime(...)` output across pages, summary cards, banners, and detail panels.
- Use a small reusable wrapper if the same `font-mono text-xs` repetition appears more than three times; otherwise inline.
- Rationale: existing table cells already use `font-mono text-xs`; the gap is everywhere else. Alternative considered was a CSS rule targeting `[data-date]` attributes, which adds markup overhead for no readability gain.

### Channel type iconography
- New component `apps/dashboard/components/channel-type-icon.tsx` exporting `<ChannelTypeIcon type={...} />` with a fixed lucide glyph per channel type:
  - `telegram` → `Send`
  - `email` → `Mail`
  - `sms` → `MessageSquare`
  - `webhook` → `Webhook`
  - `pagerduty` → `Siren`
- Rendered at `h-4 w-4` next to the type label on the channels list; type label stays text for accessibility.
- Rationale: lucide is already a dependency, no new package, consistent with sidebar icons.

### Top header removal
- Drop the `<header>` block from `AppShell` (currently lines 79-97).
- Adjust the main content wrapper's `min-h-[calc(100vh-73px)]` to `min-h-screen` since the header height is gone.
- The home page (`page.tsx`) and other pages already render their own primary `<h1>` and "Create service" CTAs in their own headers.
- Rationale: removes duplicate navigation without losing operator-visible affordances.

### Lifecycle field removal
- Drop the entire `<label>` block rendering the lifecycle value in `ServiceForm`.
- Lifecycle state still appears on the service detail summary card (`Service summary` section), on the home service health matrix, and on the services list card — all unaffected.
- Rationale: the field was informational-only; removing it reduces form noise without hiding information.

## Risks / Trade-offs

- [Top-bar removal shifts content upward by ~73px on every route] → Mitigation: pages already render their own page-level `<h1>` and CTAs, so the visual hierarchy is preserved; verify with a quick screen review before merging.
- [Mono font on large `text-xl` date labels may feel heavy] → Mitigation: limit mono to inline `text-xs` and `text-sm` date strings; keep `text-xl` "Updated" value cells as-is unless the operator reads them as data.
- [devicons-react `size` prop API may differ across icons] → Mitigation: explicit size is part of the component's typed `Props`; set it once in `ServiceIcon` and rely on the library's default styling otherwise.
- [Removing the lifecycle form field may surprise operators expecting it] → Mitigation: lifecycle state remains visible on every surface that lists services and on the service detail summary; no information loss.

## Migration Plan

- No backend changes, no data migrations.
- Pure dashboard PR: component + page edits, no infra or API work.
- Rollback: revert the PR.
