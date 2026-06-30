## ADDED Requirements

### Requirement: System exposes manual run trigger on monitor detail
System SHALL allow operators to trigger an on-demand check execution from the monitor detail view.

#### Scenario: Operator clicks manual run button
- **WHEN** operator clicks "Run check" on a monitor detail page
- **THEN** system calls `POST /api/v1/monitors/{id}/run` and shows a loading or accepted state
- **AND** after the run is accepted, operator can see the resulting run appear in the recent runs table

#### Scenario: Operator triggers run on disabled monitor
- **WHEN** operator clicks "Run check" on a monitor that is disabled
- **THEN** system shows an error message indicating the monitor must be enabled first
- **AND** no run is queued

### Requirement: System exposes incident tab on monitor detail
System SHALL show incident state for the current monitor within the monitor detail view.

#### Scenario: Monitor has open incidents
- **WHEN** operator views monitor detail and the monitor has open incidents
- **THEN** system shows a list of open incidents with incident ID, summary, opened time, and status

#### Scenario: Monitor has no incidents
- **WHEN** operator views monitor detail and the monitor has no recorded incidents
- **THEN** system shows an empty state indicating no incidents have been recorded

### Requirement: System exposes audit history tab on monitor detail
System SHALL show mutation history for the current monitor within the monitor detail view.

#### Scenario: Monitor has recorded audit events
- **WHEN** operator views monitor detail and the monitor has audit history
- **THEN** system shows a list of audit events with audit ID, event type, timestamp, and actor or origin

#### Scenario: Monitor has no audit events
- **WHEN** operator views monitor detail and the monitor has no audit history
- **THEN** system shows an empty state indicating no mutations have been recorded

### Requirement: System provides incidents overview page
System SHALL provide a page listing all incidents across monitors with filtering by status.

#### Scenario: Operator opens incidents page with no filter
- **WHEN** operator navigates to `/incidents` without a status filter
- **THEN** system returns all incidents sorted by most recent first

#### Scenario: Operator opens incidents page filtered to open
- **WHEN** operator navigates to `/incidents?status=open`
- **THEN** system returns only open and acknowledged incidents

### Requirement: System provides incident detail page
System SHALL provide a page showing full incident details with available action controls.

#### Scenario: Operator opens incident detail for open incident
- **WHEN** operator navigates to `/incidents/{id}` for an open incident
- **THEN** system shows incident ID, monitor ID, summary, opened time, current status, and available action buttons for acknowledge and resolve

#### Scenario: Operator opens incident detail for resolved incident
- **WHEN** operator navigates to `/incidents/{id}` for a resolved incident
- **THEN** system shows incident details with resolved timestamp and no action buttons

### Requirement: System provides scheduler admin page
System SHALL provide a page allowing administrators to read and update recurring execution configuration.

#### Scenario: Administrator views scheduler config
- **WHEN** administrator opens `/admin/scheduler`
- **THEN** system shows current recurring execution state and stop control mode

#### Scenario: Administrator disables recurring execution
- **WHEN** administrator submits a scheduler config update with `recurringEnabled: false`
- **THEN** system calls `PATCH /api/v1/admin/scheduler-config` with the new state and reflects the result in the UI

### Requirement: System provides probe locations page
System SHALL provide a read-only page listing all enabled probe locations.

#### Scenario: Operator opens locations page
- **WHEN** operator navigates to `/locations`
- **THEN** system shows all enabled probe locations with location ID and display name