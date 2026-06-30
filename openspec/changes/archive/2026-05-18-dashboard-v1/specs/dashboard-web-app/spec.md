## ADDED Requirements

### Requirement: System provides operator dashboard web application
System SHALL provide a web application for operators to inspect and manage monitors.

#### Scenario: Operator opens dashboard home
- **WHEN** operator navigates to the dashboard application
- **THEN** system shows a monitor-oriented overview backed by real monitor API data

### Requirement: System exposes monitor overview in dashboard
System SHALL present monitor collection data in a dashboard-friendly overview.

#### Scenario: Operator views monitor overview
- **WHEN** dashboard loads monitor collection
- **THEN** system shows each monitor with identifying configuration and available current status summary

### Requirement: System exposes monitor detail in dashboard
System SHALL provide a detailed monitor view for operational inspection.

#### Scenario: Operator views monitor detail
- **WHEN** operator opens an individual monitor
- **THEN** system shows monitor configuration, latest status, and recent run history using existing monitor read APIs

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
