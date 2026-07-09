## 1. Service Incidents API

- [x] 1.1 Add repository support for bounded service-scoped incident reads.
- [x] 1.2 Add `GET /api/v1/services/{serviceId}/incidents` handler with bounded default and maximum limits.
- [x] 1.3 Add the SST route for service-scoped incidents.
- [x] 1.4 Add API tests for existing service, missing service, empty incidents, sort order, limit behavior, and bounded query behavior.

## 2. Service Detail Header And Banner

- [x] 2.1 Remove the existing service summary card, inline edit form section, and delete service card from the main top layout.
- [x] 2.2 Add top-right outlined action buttons with icons for edit service, archive service, and create monitor.
- [x] 2.3 Add the service info banner with icon, service name, status badge, description, uptime, P99 latency, and error rate.
- [x] 2.4 Use desktop layout with identity/status on the left and indicators on the right.
- [x] 2.5 Use mobile layout that stacks the service info card content without horizontal overflow.
- [x] 2.6 Show explicit unavailable states for missing metric data.

## 3. Monitors And Alerts Layout

- [x] 3.1 Remove the monitor overview card container while keeping monitor table behavior.
- [x] 3.2 Remove probe location and protocol columns from the monitor table.
- [x] 3.3 Preserve protocol badge in the monitor name cell.
- [x] 3.4 Add recent alerts section backed by service-scoped incidents API.
- [x] 3.5 On desktop, render monitors and recent alerts in the same row with recent alerts on the right.
- [x] 3.6 On mobile, render recent alerts above monitors.
- [x] 3.7 Link recent alerts to related monitor incident context.

## 4. Danger Zone

- [x] 4.1 Add a `Danger Zone` card below operational content.
- [x] 4.2 Move delete service action and warning copy into the danger-zone card.
- [x] 4.3 Preserve existing delete confirmation and active-service blocking behavior.
- [x] 4.4 Preserve archived service read-only behavior and invalid action states.

## 5. Verification

- [x] 5.1 Run `make test-go-all`.
- [x] 5.2 Run `make lint-go`.
- [x] 5.3 Run `make lint-dashboard`.
- [x] 5.4 Run `make check-dashboard`.
- [x] 5.5 Run `make test-dashboard`.
- [x] 5.6 Run `make check-infra`.

## 6. Post-Implementation Refinements

- [x] 6.1 Add top-right outlined action buttons with icons for edit service, archive service, and create monitor (icon parity across the three).
- [x] 6.2 Color the service icon by rollup status (up/down/degraded/unknown).
- [x] 6.3 Replace inline edit form with a dedicated `Edit service` page at `/services/{serviceId}/edit`.
- [x] 6.4 Replace the monitor row action button with a kebab `More` menu (`MonitorActionsMenu`) that uses viewport-aware positioning so it stays visible on the last row.
- [x] 6.5 Render recent alert timestamps in the operator's browser timezone via a client-side formatter (`IncidentTimestamp`).
- [x] 6.6 Move destructive service delete into a name-confirmation dialog (`DeleteServiceConfirmDialog`) that requires typing the service name to enable the confirm action.
