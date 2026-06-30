## ADDED Requirements

### Requirement: Service lifecycle derives from monitor enabled count
System SHALL automatically derive service lifecycle state from the number of enabled monitors in the service.

#### Scenario: Service with zero enabled monitors derives to draft
- **WHEN** service has `enabledCount` equal to zero
- **THEN** system SHALL set service lifecycle state to `draft`

#### Scenario: Service with at least one enabled monitor derives to active
- **WHEN** service has `enabledCount` greater than zero
- **THEN** system SHALL set service lifecycle state to `active`

#### Scenario: Lifecycle re-derives when monitor is enabled
- **WHEN** monitor under service transitions from disabled to enabled
- **AND** service has `enabledCount` of zero before the transition
- **THEN** system SHALL derive service lifecycle to `active` in the same transaction

#### Scenario: Lifecycle re-derives when last monitor is disabled
- **WHEN** monitor under service transitions from enabled to disabled
- **AND** service has `enabledCount` of one before the transition
- **THEN** system SHALL derive service lifecycle to `draft` in the same transaction

#### Scenario: Lifecycle re-derives when service is created with first enabled monitor
- **WHEN** service is created with its first monitor already enabled
- **THEN** system SHALL derive service lifecycle to `active`

### Requirement: Lifecycle state is read-only in service responses
System SHALL return lifecycle state in API responses but clients SHALL NOT be able to set it directly.

#### Scenario: Lifecycle appears in service response
- **WHEN** client fetches service
- **THEN** response includes `lifecycleState` as a computed field
- **AND** field reflects current derived state based on `enabledCount`

#### Scenario: Lifecycle is derived on service read
- **WHEN** service is read after monitor enable/disable events
- **THEN** returned lifecycle reflects the most recently derived state

### Requirement: Rollup derives from monitor states for all lifecycle states
System SHALL compute service rollup status from monitor states regardless of lifecycle state.

#### Scenario: Draft service with enabled monitors shows meaningful rollup
- **WHEN** service is in `draft` lifecycle
- **AND** service has at least one enabled monitor
- **THEN** rollup status SHALL be derived from monitor check results (up/down/unknown/degraded)

#### Scenario: Archived service with enabled monitors shows meaningful rollup
- **WHEN** service is in `archived` lifecycle
- **AND** service has at least one enabled monitor
- **THEN** rollup status SHALL be derived from monitor check results

#### Scenario: Paused rollup when no enabled monitors
- **WHEN** service has `enabledCount` of zero
- **THEN** rollup status SHALL be `paused`
- **AND** this applies regardless of lifecycle state (draft/active/archived)
