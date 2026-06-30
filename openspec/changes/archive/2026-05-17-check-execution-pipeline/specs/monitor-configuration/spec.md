## MODIFIED Requirements

### Requirement: System distinguishes monitor lifecycle enablement
System SHALL track whether a monitor is enabled or disabled for execution.

#### Scenario: Disabled monitor is stored
- **WHEN** monitor configuration has disabled lifecycle state
- **THEN** system preserves that state so downstream scheduling and execution systems skip running it
