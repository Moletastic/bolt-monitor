## Why

Dashboard surfaces currently share a single generic `loading.tsx` per route segment and provide no skeleton rows inside tables. Operators see a brief empty layout before content arrives, which makes each route feel jumpy on first paint. A loading pass shapes skeletons after their destination pages and fills table bodies with placeholder rows so the layout does not collapse while data resolves.

## What Changes

- Add a `loading.tsx` to every dashboard segment that performs server data fetches. Each loading file mirrors the destination page's card grid, table column count, and primary heading so layout does not shift when content resolves.
- Inside table-based pages, render skeleton rows (`<TableRow>` with `bg-surface-low` placeholder cells) in the `<TableBody>` while data is pending, including tables that already live inside multi-fetch pages.
- Add a subtle page-level transition by introducing a per-segment `template.tsx` that wraps `children` in a fade-in animation, so navigation between routes reads as motion rather than a hard cut.
- PollingProvider (`router.refresh()`) continues to use the existing interval-driven refresh and is unaffected.

## Capabilities

### New Capabilities
- `dashboard-loading-states`: per-segment loading UI and table-level skeleton rows on dashboard surfaces.

### Modified Capabilities
- `dashboard-web-app`: extend the graceful-degradation requirement to cover the skeleton-loading pattern and the page transition.

## Impact

- `apps/dashboard/app/loading.tsx`: replace the generic 3-card layout with a per-route group of skeletons (or split into multiple loading.tsx files for distinct route shapes).
- `apps/dashboard/app/(monitoring)/services/loading.tsx`, `apps/dashboard/app/(monitoring)/integrations/channels/loading.tsx`, `apps/dashboard/app/(monitoring)/policies/loading.tsx`, `apps/dashboard/app/(monitoring)/incidents/loading.tsx`: new files.
- `apps/dashboard/components/`: add a small `Skeleton` primitive (or rely on the existing `bg-surface-low animate-pulse` pattern with a typed wrapper).
- `apps/dashboard/components/ui/table.tsx`: optional new `SkeletonRow` / `SkeletonCell` exports.
- `apps/dashboard/app/layout.tsx` and `apps/dashboard/app/(monitoring)/layout.tsx`: add a `template.tsx` that wraps `children` with the fade transition.
