## MODIFIED Requirements

### Requirement: System exposes monitor overview under services module

System SHALL provide the current monitor-oriented overview inside the `Services` module rather than on the root dashboard route.

#### Scenario: Operator opens services module

- **WHEN** operator navigates to the services landing route
- **THEN** system shows the current monitor overview backed by real monitor API data

#### Scenario: Operator opens service list with services

- **WHEN** services exist in the Services module
- **THEN** system shows each service as an actionable card that navigates to service detail from the full non-interactive card area
- **AND** each card emphasizes service name, description, rollup status, lifecycle, technology, monitor coverage, and update context
- **AND** when the service has child monitors, each card shows a per-monitor traffic-light dot row reflecting each monitor's current status
- **AND** raw service identifiers are not shown as primary card content

#### Scenario: Operator scans service technology

- **WHEN** service technology is available
- **THEN** system shows a consistently sized technology icon that is visually distinguishable in service overview, service cards, and service detail summary
- **AND** unknown or missing technology uses a consistent fallback icon with the same visual footprint

### Requirement: System provides useful integrations module state

System SHALL make notification channel management usable immediately when operators open the Integrations module.

#### Scenario: Operator opens integrations module

- **WHEN** operator navigates to `/integrations`
- **THEN** system starts loading notification channels without requiring a manual refresh action
- **AND** system shows loading feedback while the channel list is pending

#### Scenario: Notification channels are loaded

- **WHEN** notification channels are returned by the API
- **THEN** system shows configured channels with channel name, type, enabled state, default state, and available actions
- **AND** each row shows a usage-scope indicator describing how many notification routes reference the channel, with an expandable disclosure listing the referencing routes

#### Scenario: No notification channels exist

- **WHEN** notification channel API returns an empty collection
- **THEN** system shows an empty state that explains no alert channels are configured and provides a path to add one

#### Scenario: Notification channels cannot be loaded

- **WHEN** the notification channel API request fails
- **THEN** system shows an actionable unavailable state and preserves the manual retry path

## ADDED Requirements

### Requirement: System shows per-monitor traffic-light dots on the home service health matrix

The home page service health matrix SHALL render the same per-monitor traffic-light dot row on each matrix row that the services list card renders.

#### Scenario: Operator scans the home service health matrix

- **WHEN** the home page renders the service health matrix with services that have child monitors
- **THEN** each matrix row includes a per-monitor dot row matching the current status of each child monitor
- **AND** the dot row sits alongside the existing rollup status chip without replacing it

### Requirement: System marks unreferenced notification channels

The channels list SHALL distinguish referenced channels from unreferenced ones so operators can spot orphan configuration.

#### Scenario: Channel is unreferenced

- **WHEN** no notification route references a channel
- **THEN** the channel row shows an "Unused" indicator in place of the usage count
- **AND** the indicator does not link to any disclosure
