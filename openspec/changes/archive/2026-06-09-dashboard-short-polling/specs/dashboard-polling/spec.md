## ADDED Requirements

### Requirement: Dashboard provides near-real-time data updates
System SHALL automatically refresh dashboard data every 5 seconds using short polling, so users see near-real-time status without manual refresh.

#### Scenario: Data refreshes automatically
- **WHEN** user has a dashboard page open
- **THEN** system SHALL re-fetch server component data every 5 seconds
- **AND** update the displayed status, runs, and incidents without full page reload

### Requirement: Polling pauses when tab is not visible
System SHALL pause polling when the browser tab is not visible to save resources.

#### Scenario: Tab becomes hidden
- **WHEN** browser tab visibility changes to "hidden"
- **THEN** system SHALL stop the polling interval
- **AND** SHALL NOT make API calls while tab is in background

#### Scenario: Tab becomes visible
- **WHEN** browser tab visibility changes to "visible"
- **THEN** system SHALL immediately refresh data once
- **AND** resume normal 5-second polling interval

### Requirement: Polling is selective to monitoring pages
System SHALL only poll pages that display live monitoring data, not static pages like forms or settings.

#### Scenario: Services page polls
- **WHEN** user is on `/services` page
- **THEN** system SHALL poll for updated service status data

#### Scenario: Monitor detail page polls
- **WHEN** user is on `/services/:id/monitors/:id` page
- **THEN** system SHALL poll for updated runs, status, and incidents

#### Scenario: Create form does not poll
- **WHEN** user is on `/services/new` or `/services/:id/monitors/new` page
- **THEN** system SHALL NOT poll (form is static)

### Requirement: Polling uses router.refresh() for seamless updates
System SHALL use Next.js `router.refresh()` for polling to re-fetch server component data without full page navigation.

#### Scenario: Refresh occurs without navigation
- **WHEN** polling interval triggers
- **THEN** system SHALL call `router.refresh()`
- **AND** the page content updates seamlessly
- **AND** the URL and navigation state remain unchanged