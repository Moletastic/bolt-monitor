## Context

Loading feedback on the dashboard is currently inconsistent:

- `app/loading.tsx` shows three generic cards plus one large card. It applies to every route under `app/` including the home page, the services module, the integrations module, and admin.
- `app/monitors/[id]/loading.tsx` shows three generic cards but exists only for the legacy redirect route.
- No per-segment `loading.tsx` exists under `app/(monitoring)/services`, `app/(monitoring)/integrations/channels`, `app/(monitoring)/policies`, or `app/(monitoring)/incidents`.
- Tables on multi-fetch pages (home service health matrix, monitor detail runs/incidents/audit, services list, channels list, incidents list) render an empty body while data resolves.
- No route-level transition exists; navigation between segments is a hard cut.

## Goals / Non-Goals

**Goals:**
- Per-segment `loading.tsx` files for every dashboard segment that performs server data fetches.
- Skeleton rows inside table bodies on multi-fetch pages.
- A subtle per-segment page transition (fade-in) using `template.tsx`.

**Non-Goals:**
- No skeleton for client-only transitions (the polling refresh already uses the server round-trip).
- No changes to polling cadence or `PollingProvider`.
- No global loading bar / progress indicator.

## Decisions

### Per-segment `loading.tsx`
- Each segment gets a `loading.tsx` file mirroring the destination page's layout shape (count cards, summary cards, table rows).
- Skeletons use the existing `bg-surface-low animate-pulse` pattern wrapped in a small `<Skeleton>` primitive for consistency.
- No `Skeleton` exposes real text; only block placeholders.
- Rationale: keeps the visual language consistent, avoids ad-hoc class soup, and isolates per-page shape changes.

### Table loading
- Inside tables on multi-fetch pages, render `<TableRow>` placeholders inside `<TableBody>` while data is pending.
- Each placeholder row matches the destination column count (e.g., the home service health matrix has 5 columns, the monitor runs table has 6).
- The skeleton row uses variable-width `bg-surface-low` cells to avoid the visual "all-equal" look.
- Rationale: keeps the table headers stable and reduces layout shift on resolve.

### Page transition
- A `template.tsx` file in `app/` (and any deeper segments that need different timing) wraps `children` in a `motion-safe:animate-in` Tailwind animation with a short duration.
- The template re-renders on every navigation, providing a fade-in on each segment change.
- Rationale: minimal markup change, leverages Next App Router convention, no client-side router calls.

## Risks / Trade-offs

- [Adding skeletons adds JSX surface area that drifts from real pages] → Mitigation: tests assert skeleton column counts match destination tables; review per-segment shape during PR.
- [Animation may feel slow on slow connections] → Mitigation: keep duration short (150-200ms) and respect `prefers-reduced-motion`.
- [Per-segment loading.tsx proliferation] → Mitigation: keep loading components small and shared where possible (e.g., one `Loading` shell per page archetype).

## Migration Plan

- No backend, no infra, no data migrations.
- Rollback: revert the PR.
