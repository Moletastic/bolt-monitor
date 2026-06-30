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
