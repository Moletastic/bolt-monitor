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

### Requirement: Dashboard uses ServiceIcon for service technology display
System SHALL use `ServiceIcon` component with Devicon icons for displaying service technologies.

#### Scenario: Service list shows technology icons
- **WHEN** dashboard renders service list with technology information
- **THEN** each service displays its `ServiceIcon` with appropriate Devicon for the technology key

#### Scenario: ServiceIcon handles unknown technology
- **WHEN** service has an unrecognized technology key
- **THEN** `ServiceIcon` renders a fallback icon rather than text abbreviation

### Requirement: Dashboard uses MonitorProtocolBadge for monitor type display
System SHALL use `MonitorProtocolBadge` component for displaying monitor protocol types.

#### Scenario: Monitor list shows protocol badges
- **WHEN** dashboard renders monitor list with type information
- **THEN** each monitor displays its `MonitorProtocolBadge` with styled text (HTTP, HTTPS, TCP, gRPC, DNS)

#### Scenario: MonitorProtocolBadge is text-only
- **WHEN** `MonitorProtocolBadge` renders any protocol
- **THEN** badge contains only styled text, no icon elements

### Requirement: Dashboard sidebar uses Lucide icons
System SHALL render sidebar navigation items with Lucide icons.

#### Scenario: Sidebar nav items have icons
- **WHEN** dashboard renders sidebar navigation
- **THEN** each nav item (Services, Monitors, Incidents, Settings) includes a Lucide icon alongside its label
