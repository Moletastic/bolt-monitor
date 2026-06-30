## MODIFIED Requirements

### Requirement: System exposes monitor overview with protocol context
System SHALL show protocol/type context when listing monitors.

#### Scenario: Operator scans monitor overview on desktop
- **WHEN** operator views a monitor overview table
- **THEN** system includes a dedicated protocol/type column separate from monitor name
- **AND** rows show monitor name, protocol/type, current status, enabled state, last check, duration, and available action
- **AND** raw monitor identifiers are not shown as primary row content

### Requirement: System exposes monitor detail in dashboard
System SHALL provide a detailed monitor view for operational inspection.

#### Scenario: Operator reviews current monitor status
- **WHEN** operator opens monitor detail
- **THEN** system shows a current-status summary with monitor name, protocol/type, target, enabled state, current status, last outcome, last check time, duration, and cadence
- **AND** system shows the latest error when status data includes an error
- **AND** raw service or monitor identifiers are not shown as primary status content

### Requirement: System exposes settings module overview
System SHALL provide a settings module overview for dashboard control-plane context.

#### Scenario: Operator opens settings module
- **WHEN** operator navigates to `/config`
- **THEN** system shows a settings overview instead of placeholder content
- **AND** the overview includes scheduler recurring execution state and safe setup/environment context

#### Scenario: Settings source data is unavailable
- **WHEN** scheduler configuration data cannot be loaded
- **THEN** system shows an unavailable state for the affected settings section while preserving the rest of the settings page

### Requirement: System supports monitor management through dashboard
The dashboard SHALL allow operators to manage monitor configuration from the dashboard.

#### Scenario: Operator creates monitor from dashboard
- **WHEN** operator submits a valid create-monitor form
- **THEN** system creates the monitor through the existing monitor create API and reflects the new monitor in dashboard views
- **AND** the submitted payload does not include probe-location or region selection

#### Scenario: Operator updates monitor from dashboard
- **WHEN** operator submits valid monitor changes
- **THEN** system updates the monitor through the existing monitor update API and reflects the saved state in dashboard views
- **AND** the submitted payload does not include probe-location or region selection

## REMOVED Requirements

### Requirement: System presents probe-location selection honestly
The dashboard SHALL derive any monitor probe-location selection from the enabled subset of the canonical probe-location catalog read from the monitor API at request time. The dashboard SHALL NOT hard-code probe-location identifiers in client components or server actions as the source of selection.

#### Scenario: Catalog contains a single enabled location
- **WHEN** the enabled subset of the probe-location catalog contains exactly one location
- **THEN** the monitor form renders a non-interactive region chip showing the location name and a helper text indicating single-region preview
- **AND** the create and update monitor server actions submit the location from the catalog data, not from a constant

#### Scenario: Catalog contains multiple enabled locations
- **WHEN** the enabled subset of the probe-location catalog contains more than one location
- **THEN** the monitor form renders a real selection control bound to the enabled locations
- **AND** the dashboard does not impose a single-selection default that is not present in the catalog
