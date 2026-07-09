## Why

The service detail page currently spends too much space on summary, edit, and delete cards, which pushes operational data lower and mixes routine actions with destructive actions. Operators need a denser service overview, faster primary actions, recent alert context, and a safer danger-zone layout.

## What Changes

- Replace the current service summary, edit service, and delete service sections with top-right outlined action buttons for edit service, archive service, and create monitor.
- Move edit-service into a dedicated `Edit service` page at `/services/{serviceId}/edit` so the detail page stays focused on operational state.
- Add a full-width service info banner with service icon, service name, status badge, description, and three metrics: uptime, P99 latency, and error rate.
- Color the service icon by rollup status (up/down/degraded/unknown) so operators see service health at a glance.
- Use responsive layout: desktop keeps service identity/status on the left and metrics on the right; mobile stacks the info card content.
- Remove the monitor overview card container and render the monitor table as the primary monitor surface.
- Remove probe location and protocol columns from the monitor table; keep protocol badge inside the monitor name cell.
- Move monitor row actions into a kebab `More` menu with viewport-aware positioning so the menu stays visible on the last row.
- Add a recent alerts section based on service-level incident data for monitors under the service.
- Use desktop layout with monitors and recent alerts in the same row, with recent alerts on the right; mobile stacks recent alerts above monitors.
- Render recent alert timestamps in the operator's browser timezone via a client-side formatter.
- Add a danger-zone card containing the delete service action, warning copy, and an inline active-service block warning.
- Move the destructive service delete into a name-confirmation dialog (operator must type the service name to enable the confirm action); replace the prior generic `ConfirmDialog` for this destructive flow.
- Add or extend a backend endpoint to fetch recent service incidents without listing unrelated incident data.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: Refactor service detail layout, action placement, monitor table columns, recent alerts, and danger zone.
- `incident-management-api`: Add service-scoped incident read behavior for recent alerts on service detail.

## Impact

- Affected dashboard route: `apps/dashboard/app/(monitoring)/services/[serviceId]/page.tsx`.
- Affected dashboard components: monitor table, monitor actions menu, service action controls, service icon, service summary/banner components, recent alerts, danger-zone delete UI, edit-service page.
- New dashboard route: `apps/dashboard/app/services/[serviceId]/edit/page.tsx`.
- Affected API route: service-level incident read endpoint, `GET /api/v1/services/{serviceId}/incidents?limit=<n>`.
- Affected repository: efficient lookup of incidents for monitors belonging to a service.
- No service, monitor, or incident mutation behavior changes are expected.
