## MODIFIED Requirements

### Requirement: System provides operator dashboard web application
System SHALL provide a web application for operators to inspect and manage services and nested monitors through a module-oriented console layout.

#### Scenario: Operator opens dashboard home
- **WHEN** operator navigates to the dashboard application
- **THEN** system shows a dashboard landing page framed inside the shared dashboard sidebar shell
- **AND** that landing page can use explicit work-in-progress or empty-state messaging while broader dashboard content is not yet implemented

### Requirement: System exposes monitor overview under services module
System SHALL provide a real service-oriented overview inside the `Services` module rather than a monitor-oriented overview on the root dashboard route.

#### Scenario: Operator opens services module
- **WHEN** operator navigates to the services landing route
- **THEN** system shows service summaries backed by real service and nested monitor API data

### Requirement: System exposes monitor detail in dashboard
System SHALL provide a detailed nested monitor view for operational inspection inside the services module.

#### Scenario: Operator views monitor detail
- **WHEN** operator opens an individual monitor from its parent service context
- **THEN** system shows monitor configuration, latest status, and recent run history using nested monitor read APIs
- **AND** keeps the monitor detail view inside the same module-oriented dashboard shell with `Services` treated as the active module

### Requirement: System supports monitor management through dashboard
System SHALL allow operators to manage services and nested monitor configuration from the dashboard.

#### Scenario: Operator creates service from dashboard
- **WHEN** operator submits a valid create-service form
- **THEN** system creates the service through the service create API and reflects the new service in dashboard views

#### Scenario: Operator creates monitor from dashboard
- **WHEN** operator submits a valid create-monitor form under existing service
- **THEN** system creates the nested monitor through the nested monitor create API and reflects the new monitor in dashboard views

#### Scenario: Operator updates monitor from dashboard
- **WHEN** operator submits valid nested monitor changes
- **THEN** system updates the monitor through the nested monitor update API and reflects the saved state in dashboard views

#### Scenario: Operator enables or disables monitor from dashboard
- **WHEN** operator triggers enable or disable control for a nested monitor
- **THEN** system calls the nested action endpoint and reflects the resulting enabled state in dashboard views

### Requirement: System preserves monitoring design language in dashboard
System SHALL implement dashboard UI using the repository's monitoring design system.

#### Scenario: Dashboard renders status-oriented UI
- **WHEN** system renders dashboard surfaces
- **THEN** typography, colors, spacing, density, and status emphasis follow `DESIGN.md` and remain consistent with the intended monitoring-console visual language
