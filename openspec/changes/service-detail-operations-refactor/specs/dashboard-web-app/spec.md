## ADDED Requirements

### Requirement: System presents service detail as operational overview

The dashboard SHALL present service detail as an operational overview with top actions, service identity, service health metrics, monitor list, recent alerts, and danger zone.

#### Scenario: Operator opens service detail on desktop
- **WHEN** the operator opens a service detail page on a wide viewport
- **THEN** system shows outlined edit service, archive service, and create monitor actions at the top-right of the page
- **AND** system shows a full-width service info banner below the actions
- **AND** the banner places service icon, service name, status badge, and description on the left
- **AND** the banner places uptime, P99 latency, and error rate indicators on the right

#### Scenario: Operator opens service detail on mobile
- **WHEN** the operator opens a service detail page on a narrow viewport
- **THEN** system stacks actions, service info card, recent alerts, monitors, and danger zone in that order
- **AND** the service info card remains readable without horizontal scrolling

#### Scenario: Service metrics are unavailable
- **WHEN** uptime, P99 latency, or error-rate data is unavailable
- **THEN** system shows an explicit unavailable value rather than implying healthy zero or perfect values

#### Scenario: Operator reviews service actions
- **WHEN** the service is not archived
- **THEN** system provides outlined actions to edit the service, archive the service, and create a monitor
- **WHEN** the service is archived
- **THEN** system disables or hides actions that are invalid for archived services
- **AND** system preserves read-only service context

### Requirement: System simplifies service monitor list

The dashboard SHALL render service monitors without redundant overview chrome or deprecated columns.

#### Scenario: Operator views monitors on service detail
- **WHEN** monitors exist for a service
- **THEN** system renders the monitor table without a surrounding monitor overview card container
- **AND** system omits the probe location column
- **AND** system omits the protocol column
- **AND** the monitor name cell continues to include the protocol badge before the monitor name

#### Scenario: Service has no monitors
- **WHEN** the service has no monitors
- **THEN** system preserves an empty state that guides valid monitor creation for non-archived services
- **AND** system does not offer monitor creation for archived services

### Requirement: System shows recent service alerts

The dashboard SHALL show recent alerts for incidents related to monitors under the current service.

#### Scenario: Service has recent alerts on desktop
- **WHEN** service-level incident data includes recent incidents and the viewport is wide
- **THEN** system shows recent alerts in the same row as the monitor list
- **AND** recent alerts appear to the right of the monitor list
- **AND** each alert links to the related monitor incident context

#### Scenario: Service has recent alerts on mobile
- **WHEN** service-level incident data includes recent incidents and the viewport is narrow
- **THEN** system shows recent alerts above the monitor list
- **AND** each alert remains tappable and readable

#### Scenario: Service has no recent alerts
- **WHEN** service-level incident data is empty
- **THEN** system shows a low-emphasis empty recent alerts state

### Requirement: System isolates destructive service deletion

The dashboard SHALL isolate permanent service deletion in a danger-zone card.

#### Scenario: Operator views danger zone
- **WHEN** the operator opens service detail
- **THEN** system shows a `Danger Zone` card below operational content
- **AND** the card explains that deleting a service is permanent
- **AND** the delete service action is only available inside the danger-zone card
- **AND** existing delete confirmation and active-service blocking behavior are preserved

### Requirement: System requires typing the service name to confirm deletion

The dashboard SHALL require the operator to type the service name verbatim before the destructive delete action can be submitted.

#### Scenario: Operator opens the delete service dialog
- **WHEN** the operator clicks `Delete service`
- **THEN** system opens a confirmation modal showing the service name
- **AND** the confirm action button stays disabled until the typed value matches `service.name` exactly (whitespace trimmed)
- **AND** the modal has a close (X) button at the top-right and a `Cancel` action
- **AND** submitting the form routes through the existing `deleteServiceAction` server action

#### Scenario: Service is archived or has active monitors
- **WHEN** the service cannot currently be deleted
- **THEN** the trigger `Delete service` button is disabled
- **AND** the confirmation modal does not open

### Requirement: System routes service edits through a dedicated edit page

The dashboard SHALL route the `Edit service` action to a dedicated edit page rather than an inline form on the detail page.

#### Scenario: Operator opens the edit service page
- **WHEN** the operator clicks `Edit service` on a non-archived service
- **THEN** system navigates to `/services/{serviceId}/edit`
- **AND** the page reuses the existing `ServiceForm` in edit mode bound to the current service

#### Scenario: Operator opens the edit service page for an archived service
- **WHEN** the service is archived
- **THEN** system renders a read-only notice with a link back to the detail page
- **AND** the editable `ServiceForm` is not rendered

### Requirement: System exposes monitor row actions through a viewport-aware kebab menu

The dashboard SHALL expose per-monitor row actions through a kebab `More` menu that flips open direction when there is not enough space below the trigger.

#### Scenario: Operator opens the monitor actions menu on a middle row
- **WHEN** the operator clicks the kebab trigger on any monitor row that has viewport space below
- **THEN** the menu opens below the trigger with the available actions listed as `role="menuitem"` items

#### Scenario: Operator opens the monitor actions menu on the last row
- **WHEN** the operator clicks the kebab trigger on the last visible monitor row and there is insufficient viewport space below
- **THEN** the menu opens above the trigger so it is fully visible
- **AND** the flip direction recomputes on `resize` and on scroll within any ancestor scroll container

### Requirement: System renders recent alert timestamps in the operator's timezone

The dashboard SHALL render recent alert timestamps using the operator's browser timezone.

#### Scenario: Operator reviews recent alerts
- **WHEN** the operator views the recent alerts section
- **THEN** each incident timestamp reflects the operator's local timezone via `Intl.DateTimeFormat` evaluated on the client
- **AND** the timestamp is rendered through a dedicated client component so SSR does not lock the value to the server timezone
