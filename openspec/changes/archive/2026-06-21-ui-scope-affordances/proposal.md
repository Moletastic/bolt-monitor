## Why

Operators triaging a service or a notification channel need to see scope at a glance. Today a service card shows only the rollup status, hiding which child monitors are up, degraded, or down. A notification channel row shows only the channel metadata, hiding which notification routes use it. Both surfaces gain operational clarity from a small, static indicator of where they sit in the wider system.

## What Changes

- Render a static row of per-monitor traffic-light dots on each service card in the services list (`/services`) and the home service health matrix. Each dot reflects the current status of one child monitor (`UP`, `DOWN`, `DEGRADED`, `MAINTENANCE`, `UNKNOWN`). Dots are informational only — no click target — and a `title` attribute carries the monitor name for screen reader and hover affordance.
- Render a "Used by N routes" link on each notification channels list row, computed server-side from the existing escalation policies list. Clicking the link opens an in-page disclosure listing the routes that reference the channel.
- No backend data model changes. Both new affordances consume data already returned by existing APIs (`listServices`, `getService`, `listEscalationPolicies`).

## Capabilities

### New Capabilities
- `dashboard-scope-affordances`: per-monitor traffic-light pills on service surfaces and channel usage scope on the channels list.

### Modified Capabilities
- `dashboard-web-app`: extend the service card requirement to include a per-monitor traffic-light row, and extend the integrations/channels requirement to include a usage-scope disclosure.

## Impact

- `apps/dashboard/app/(monitoring)/services/page.tsx`: render traffic-light dots inside each service card.
- `apps/dashboard/app/page.tsx`: render traffic-light dots inside the home service health matrix rows.
- `apps/dashboard/components/`: add a `monitor-traffic-light.tsx` (or inline dot list) component.
- `apps/dashboard/app/(monitoring)/integrations/channels/page.tsx`: fetch escalation policies in parallel with the channel list, compute usage map, render the "Used by N routes" disclosure.
- `apps/dashboard/components/`: add a `channel-usage-scope.tsx` disclosure component.
