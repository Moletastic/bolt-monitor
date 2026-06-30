## ADDED Requirements

### Requirement: System derives service rollup status from enabled child monitors
System SHALL derive service rollup status from service lifecycle state and enabled child monitor states.

#### Scenario: Draft service reports draft rollup
- **WHEN** service lifecycle state is `draft`
- **THEN** derived service rollup status is `draft`

#### Scenario: Active service with all monitors disabled reports paused rollup
- **WHEN** service lifecycle state is `active`
- **AND** all child monitors are disabled
- **THEN** derived service rollup status is `paused`

#### Scenario: Active service with no observed enabled monitor status reports unknown rollup
- **WHEN** service lifecycle state is `active`
- **AND** at least one child monitor is enabled
- **AND** no enabled child monitor has current status yet
- **THEN** derived service rollup status is `unknown`

#### Scenario: Active service with all enabled monitors up reports up rollup
- **WHEN** service lifecycle state is `active`
- **AND** all enabled child monitors are `up`
- **THEN** derived service rollup status is `up`

#### Scenario: Active service with all enabled monitors down reports down rollup
- **WHEN** service lifecycle state is `active`
- **AND** all enabled child monitors are `down`
- **THEN** derived service rollup status is `down`

#### Scenario: Active service with mixed enabled monitor states reports degraded rollup
- **WHEN** service lifecycle state is `active`
- **AND** enabled child monitors contain both `up` and `down` states
- **THEN** derived service rollup status is `degraded`

### Requirement: System ignores disabled monitors in service rollup derivation
System SHALL ignore disabled child monitors when deriving current service rollup state.

#### Scenario: Disabled monitor does not affect service rollup
- **WHEN** active service contains disabled child monitor with stale or down status
- **THEN** system ignores that disabled child monitor when deriving service rollup status

### Requirement: System prevents active services from entering invalid zero-monitor operating state
System SHALL prevent behaviors that leave an active service with zero monitors.

#### Scenario: Client deletes last monitor from active service
- **WHEN** client attempts to delete the last remaining monitor under active service
- **THEN** system rejects request without deleting that monitor
- **AND** system preserves existing service lifecycle state
