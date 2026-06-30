## MODIFIED Requirements

### Requirement: System provides operator dashboard web application
System SHALL provide a web application for operators to inspect monitoring health, triage operational attention, and manage monitors through a module-oriented console layout.

#### Scenario: Operator opens dashboard home
- **WHEN** operator navigates to the dashboard application
- **THEN** system shows an operational overview framed inside the shared dashboard sidebar shell
- **AND** the overview summarizes service health, incident state, scheduler state, and setup gaps using available dashboard APIs

#### Scenario: Operator sees prioritized attention
- **WHEN** operator opens the dashboard home and there are down services, open incidents, disabled scheduler state, services without monitors, disabled monitor coverage, or draft services
- **THEN** system shows a prioritized attention area that identifies the items needing operator review
- **AND** each actionable item links to the existing module route where the operator can inspect or manage it

#### Scenario: Operator reviews service health matrix
- **WHEN** operator opens the dashboard home and services exist
- **THEN** system shows a compact service health matrix with service identity, rollup status, lifecycle state, monitor coverage, and recent update context
- **AND** each service row links to the existing service detail route

#### Scenario: Operator opens dashboard with no configured services
- **WHEN** operator navigates to the dashboard application and no services exist
- **THEN** system shows an empty-state path to create the first service
- **AND** system does not show misleading zero-health summaries as if monitoring coverage exists

#### Scenario: Dashboard overview API context is unavailable
- **WHEN** one or more dashboard overview API requests fail
- **THEN** system shows an actionable unavailable or partial-state message for the affected overview content
- **AND** system preserves the shared dashboard shell and navigation
