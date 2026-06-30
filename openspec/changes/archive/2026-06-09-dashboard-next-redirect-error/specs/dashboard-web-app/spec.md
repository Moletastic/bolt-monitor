## MODIFIED Requirements

### Requirement: System provides operator dashboard web application
System SHALL provide a web application for operators to inspect and manage monitors through a module-oriented console layout.

#### Scenario: Operator opens dashboard home
- **WHEN** operator navigates to the dashboard application
- **THEN** system shows a dashboard landing page framed inside the shared dashboard sidebar shell
- **AND** that landing page can use explicit work-in-progress or empty-state messaging while broader dashboard content is not yet implemented

### Requirement: System exposes monitor overview under services module
System SHALL provide the current monitor-oriented overview inside the `Services` module rather than on the root dashboard route.

#### Scenario: Operator opens services module
- **WHEN** operator navigates to the services landing route
- **THEN** system shows the current monitor overview backed by real monitor API data

### Requirement: System exposes monitor detail in dashboard
System SHALL provide a detailed monitor view for operational inspection.

#### Scenario: Operator views monitor detail
- **WHEN** operator opens an individual monitor
- **THEN** system shows monitor configuration, latest status, and recent run history using existing monitor read APIs
- **AND** keeps the monitor detail view inside the same module-oriented dashboard shell with `Services` treated as the active module

### Requirement: System supports monitor management through dashboard
System SHALL allow operators to manage monitor configuration from the dashboard.

#### Scenario: Operator creates monitor from dashboard
- **WHEN** operator submits a valid create-monitor form
- **THEN** system creates the monitor through the existing monitor create API and reflects the new monitor in dashboard views

#### Scenario: Operator updates monitor from dashboard
- **WHEN** operator submits valid monitor changes
- **THEN** system updates the monitor through the existing monitor update API and reflects the saved state in dashboard views

#### Scenario: Operator enables or disables monitor from dashboard
- **WHEN** operator triggers enable or disable control for a monitor
- **THEN** system calls the existing action endpoint and reflects the resulting enabled state in dashboard views

### Requirement: System preserves monitoring design language in dashboard
System SHALL implement dashboard UI using the repository's monitoring design system.

#### Scenario: Dashboard renders status-oriented UI
- **WHEN** system renders dashboard surfaces
- **THEN** typography, colors, spacing, density, and status emphasis follow `DESIGN.md` and remain consistent with the intended monitoring-console visual language

### Requirement: System handles form submission redirects correctly
System SHALL use proper Next.js redirect patterns in server actions and form submissions without throwing unhandled errors.

#### Scenario: Form submit redirects without error
- **WHEN** operator submits a form that requires redirection after success
- **THEN** system SHALL use `redirect()` from `next/navigation` as a return value
- **AND** SHALL NOT wrap `redirect()` in try-catch blocks that would catch NEXT_REDIRECT as an error

#### Scenario: Form submit with validation error does not redirect
- **WHEN** operator submits a form with invalid data
- **THEN** system SHALL return a validation error response
- **AND** SHALL NOT call `redirect()`
