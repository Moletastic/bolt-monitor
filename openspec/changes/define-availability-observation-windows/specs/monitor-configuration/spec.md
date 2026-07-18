## ADDED Requirements

### Requirement: System supports optional monitor availability objectives
System SHALL allow monitor configuration to omit an availability objective or reference one valid immutable effective-dated objective definition using integer target basis points.

#### Scenario: Monitor is configured without objective
- **WHEN** a client creates or updates a monitor without objective configuration
- **THEN** validation accepts the monitor when all existing required fields are valid

#### Scenario: Monitor objective is configured
- **WHEN** a client supplies objective type, integer `targetBasisPoints`, rolling duration, and any type-required threshold
- **THEN** the canonical monitor contract preserves the objective reference for API, persistence, scheduling, and reporting consumers

### Requirement: Scheduling and enabled-state history is immutable and enumerable
System SHALL append immutable effective-dated schedule definitions and enabled/disabled intervals for monitor scheduling changes, including UTC interval alignment sufficient to enumerate historical expected slots independently of scheduler work.

#### Scenario: Monitor cadence changes
- **WHEN** a monitor's recurring interval changes
- **THEN** the prior schedule definition ends and a new non-overlapping version begins at a recurring slot boundary
- **AND** both versions remain available for finalization, reporting, and recovery validation

#### Scenario: Monitor enablement changes
- **WHEN** a monitor is enabled or disabled
- **THEN** the system appends an immutable half-open enabled/disabled interval fact with actor/source, reason, recorded time, effective bounds, version, and audit identity
- **AND** the transition is not inferred later from mutable current monitor state
