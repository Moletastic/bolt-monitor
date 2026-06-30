## 1. Per-monitor traffic-light component

- [x] 1.1 Create `apps/dashboard/components/monitor-traffic-light.tsx` exporting `<MonitorTrafficLights monitors={...} maxVisible={12} />`.
- [x] 1.2 Implement dot color mapping per status (`UP` → `bg-status-up`, `DEGRADED` → `bg-status-warn`, `DOWN` → `bg-status-down`, `MAINTENANCE` → `bg-muted-foreground`, unknown → `bg-muted-foreground/50`).
- [x] 1.3 Add `title` and `aria-label` per dot carrying the monitor name and status.
- [x] 1.4 Cap the visible dot row and append a `+N` indicator when monitor count exceeds `maxVisible`.

## 2. Traffic-light dots on service cards

- [x] 2.1 Update `apps/dashboard/app/(monitoring)/services/page.tsx` to render `<MonitorTrafficLights monitors={service.monitors ?? []} />` inside each service card below the status/lifecycle row.
- [x] 2.2 Update `apps/dashboard/app/page.tsx` to render the same component inside each row of `ServiceHealthMatrix`.

## 3. Channel usage scope

- [x] 3.1 Update `apps/dashboard/app/(monitoring)/integrations/channels/page.tsx` to fetch escalation policies in parallel with the channel list (`Promise.all`).
- [x] 3.2 Compute a usage map keyed by `channelId` by walking each policy's `businessHoursPath.steps` and `offHoursPath.steps`.
- [x] 3.3 Add a `<ChannelUsageScope channelId={...} policies={...} />` disclosure component (server-renderable using `<details>`/`<summary>` or a small client island).
- [x] 3.4 Render "Used by N routes" inside each channels list row; render "Unused" when the count is zero.

## 4. Verification

- [x] 4.1 Run `make check-dashboard` and `make lint-dashboard`.
- [x] 4.2 Run `make build-dashboard`.
