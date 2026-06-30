# monitor-minute-cadence-validation Specification

## Purpose
TBD - created by archiving change per-monitor-interval-execution. Update Purpose after archive.
## Requirements
### Requirement: Monitor cadence uses supported minute-based presets
System SHALL accept only supported minute-based `intervalSeconds` values for monitor configuration.

#### Scenario: Supported cadence accepted
- **WHEN** a monitor is created or updated with `intervalSeconds` equal to 60, 120, 180, 300, 600, 900, 1800, or 3600
- **THEN** validation SHALL accept the cadence

#### Scenario: Unsupported cadence rejected
- **WHEN** a monitor is created or updated with `intervalSeconds` equal to an unsupported value such as 30, 90, or 150
- **THEN** validation SHALL reject the monitor configuration

#### Scenario: User-facing cadence labels are minute based
- **WHEN** the dashboard presents cadence choices
- **THEN** it SHALL label choices as minutes or hours instead of raw seconds
