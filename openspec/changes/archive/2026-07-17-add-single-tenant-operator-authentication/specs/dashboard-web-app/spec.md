## MODIFIED Requirements

### Requirement: System provides operator dashboard web application
System SHALL provide a web application for authenticated operators whose authoritative membership status is `ACTIVE` to inspect monitoring health, triage operational attention, and manage monitors through a module-oriented console layout. Existing operator surfaces SHALL render only after server-side dashboard session validation; custom authentication and recovery pages SHALL render outside the protected operator shell.

#### Scenario: Unauthenticated visitor opens dashboard home
- **WHEN** a visitor without a valid dashboard session navigates to the dashboard application
- **THEN** the server redirects to the custom sign-in page before rendering operational data or the protected dashboard shell

#### Scenario: Non-active operator opens dashboard home
- **WHEN** an authenticated dashboard session receives membership-denied feedback from the monitor API
- **THEN** the dashboard invalidates the session and requires authentication again
- **AND** it does not render protected operational data from that request

#### Scenario: Operator opens dashboard home
- **WHEN** an authenticated `ACTIVE` operator navigates to the dashboard application
- **THEN** system shows an operational overview framed inside the shared dashboard sidebar shell
- **AND** the overview summarizes service health, incident state, scheduler state, and setup gaps using available dashboard APIs

#### Scenario: Operator sees prioritized attention
- **WHEN** an authenticated `ACTIVE` operator opens the dashboard home and there are down services, open incidents, disabled scheduler state, services without monitors, disabled monitor coverage, or draft services
- **THEN** system shows a prioritized attention area that identifies the items needing operator review
- **AND** each actionable item links to the existing module route where the operator can inspect or manage it

#### Scenario: Operator reviews service health matrix
- **WHEN** an authenticated `ACTIVE` operator opens the dashboard home and services exist
- **THEN** system shows a compact service health matrix with service identity, rollup status, lifecycle state, monitor coverage, and recent update context
- **AND** each service row links to the existing service detail route

#### Scenario: Operator opens dashboard with no configured services
- **WHEN** an authenticated `ACTIVE` operator navigates to the dashboard application and no services exist
- **THEN** system shows an empty-state path to create the first service
- **AND** system does not show misleading zero-health summaries as if monitoring coverage exists

#### Scenario: Dashboard overview API context is unavailable
- **WHEN** one or more dashboard overview API requests fail for an authenticated `ACTIVE` operator
- **THEN** system shows an actionable unavailable or partial-state message for the affected overview content
- **AND** system preserves the shared dashboard shell and navigation

## ADDED Requirements

### Requirement: Authentication pages use a separate public dashboard segment
The dashboard SHALL place sign-in, invitation activation, password recovery, password reset, and TOTP challenge or enrollment pages in an explicitly public route segment that does not instantiate polling, protected navigation, or operational API reads.

#### Scenario: Visitor opens sign-in
- **WHEN** an unauthenticated visitor opens the custom sign-in route
- **THEN** the page renders without the operator sidebar, polling provider, or a request to `/api/v1/**`

#### Scenario: Authentication page renders an error
- **WHEN** an authentication operation returns safe operator feedback
- **THEN** the page preserves the authentication form context without exposing Cognito internals, submitted secrets, or raw exception text
