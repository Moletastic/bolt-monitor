## 1. Skeleton primitive

- [x] 1.1 Add `apps/dashboard/components/ui/skeleton.tsx` exporting a typed `<Skeleton>` block using `bg-surface-low animate-pulse rounded-md`.
- [x] 1.2 Verify the primitive renders with consistent height/width tokens across uses.

## 2. Per-segment loading.tsx

- [x] 2.1 Add `apps/dashboard/app/(monitoring)/services/loading.tsx` mirroring the services list card grid and table shape.
- [x] 2.2 Add `apps/dashboard/app/(monitoring)/integrations/channels/loading.tsx` mirroring the channels table.
- [x] 2.3 Add `apps/dashboard/app/(monitoring)/policies/loading.tsx` mirroring the policy card grid.
- [x] 2.4 Add `apps/dashboard/app/(monitoring)/incidents/loading.tsx` mirroring the incident table.
- [x] 2.5 Update `apps/dashboard/app/loading.tsx` so the home page skeleton mirrors the home page layout (attention queue card + service health matrix + recent incidents + setup gaps).
- [x] 2.6 Update `apps/dashboard/app/monitors/[id]/loading.tsx` to match the monitor detail skeleton (status card + tabs + form + delete card).

## 3. Table skeleton rows

- [x] 3.1 Update `apps/dashboard/components/ui/table.tsx` (or a sibling file) to export `<TableRowSkeleton columns={N} />` and `<TableCellSkeleton />` for reuse.
- [x] 3.2 Wire skeleton rows into the home service health matrix, monitor detail (runs/incidents/audit), services list, channels list, and incidents list while data is in flight.

## 4. Page transition

- [x] 4.1 Add `apps/dashboard/app/template.tsx` that wraps `children` with a Tailwind fade-in animation honoring `motion-safe:` and `prefers-reduced-motion`.
- [x] 4.2 Add a `template.tsx` to `apps/dashboard/app/(monitoring)/` if a different timing is desired for monitoring routes.
- [x] 4.3 Verify animation runs on segment navigation and is suppressed when reduced motion is requested.

## 5. Verification

- [x] 5.1 Run `make check-dashboard` and `make lint-dashboard`.
- [x] 5.2 Run `make build-dashboard`.
