## Context

Operators scan service cards and channel rows looking for where to spend attention. Two affordances improve scanning density without changing product behavior:

1. **Service cards**: today the service card surfaces the rollup status only. The list API already returns `service.monitors[]` with `status.currentStatus` populated, so per-monitor dots are derivable from data already in hand.
2. **Channels list**: today each row shows name, type, target, updated time. There is no signal showing whether a channel is orphaned (configured but unused) or central to several routes. The escalation policies API returns the full path/step data needed to compute usage server-side.

## Goals / Non-Goals

**Goals:**
- Add per-monitor traffic-light dots to service cards and the home service health matrix.
- Add a "Used by N routes" disclosure to each channels list row.
- Keep both affordances informational and low-noise — no new click targets beyond the existing disclosure control.

**Non-Goals:**
- No click-through drill-down from individual dots.
- No backend schema changes (no `usageCount` field added to channel payload).
- No change to the channels list primary action (still links from the channel name to the channel detail page).

## Decisions

### Per-monitor traffic-light dots

- New component `<MonitorTrafficLights monitors={...} />` renders one small dot per monitor using the existing status color tokens:
  - `UP` → `bg-status-up`
  - `DEGRADED` → `bg-status-warn`
  - `DOWN` → `bg-status-down`
  - `MAINTENANCE` → `bg-muted-foreground`
  - `UNKNOWN` / undefined → `bg-muted-foreground/50`
- Each dot is `h-2 w-2 rounded-full` with `title={monitor.name}` and `aria-label={monitor.name + ': ' + status}`.
- Layout: a flex row with `gap-1`, wrapped to multiple lines if monitor count exceeds a comfortable row width.
- Dots are not wrapped in `<Link>`; the surrounding service card link remains the navigation target.
- Rationale: data already on hand; visual signal is unmistakable; tooltip gives monitor name on hover.

### Channel usage scope

- The channels list page server component fetches escalation policies in parallel with channels using `Promise.all`.
- A usage map is built by walking each policy's `businessHoursPath.steps[].channelId` and `offHoursPath.steps[].channelId`, counting references per channel.
- Each row renders a "Used by N routes" link that toggles an in-page disclosure (`<details>`/`<summary>` is acceptable, or a small client island with `useState`).
- Disclosure lists the routes that reference the channel by name, each rendered as a `<Link href="/policies/{policyId}">`.
- Channels with zero references show "Unused" with no link target.
- Rationale: server-side aggregation avoids a backend payload change; `<details>` keeps it server-renderable without client state; the disclosure is the only new interactive element.

## Risks / Trade-offs

- [Per-monitor dots add visual noise to cards with many monitors] → Mitigation: cap the visible dot row to ~12 monitors and append "+N" if longer.
- [Channel usage disclosure may be confusing when a channel is used only by an archived service] → Mitigation: show usage count by route only, not by service; out-of-scope services are filtered out by route listing.
- [Two parallel API fetches on the channels page increase load time slightly] → Mitigation: both fetches already exist as parallel reads elsewhere on the dashboard; the new pair adds at most one round-trip.

## Migration Plan

- No backend migrations, no infra changes.
- Rollback: revert the PR.
