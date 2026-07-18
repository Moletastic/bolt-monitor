## ADDED Requirements

### Requirement: Maintenance exit awaits a new recurring observation
System SHALL represent a monitor leaving maintenance as `UNKNOWN` with an awaiting-observation reason until a new recurring canonical result establishes current state.

#### Scenario: Maintenance ends after prior up status
- **WHEN** a maintenance interval ends for a monitor whose pre-maintenance status was `UP`
- **THEN** latest status becomes `UNKNOWN` and indicates awaiting recurring observation
- **AND** the system does not restore `UP` from the stale pre-maintenance snapshot

#### Scenario: Manual check succeeds after maintenance
- **WHEN** a manual check succeeds while the monitor is awaiting recurring observation
- **THEN** the manual result remains visible in run history
- **AND** it does not clear awaiting-observation state by default

#### Scenario: Recurring check completes after maintenance
- **WHEN** the first post-maintenance recurring canonical result is processed
- **THEN** current monitor state is derived from that new result and existing state-machine thresholds
- **AND** awaiting-observation reason is cleared

### Requirement: Objective reads do not replace operational status reads
System SHALL keep availability objective reporting separate from latest status and recent raw execution history.

#### Scenario: Monitor is currently up but objective is incomplete
- **WHEN** latest recurring status is `UP` and the objective window contains missing slots
- **THEN** status reads continue to report current `UP` state
- **AND** availability reads independently report `INCOMPLETE`

#### Scenario: Objective finalizer is behind
- **WHEN** latest status is current but the independent objective-finalizer watermark is stale or has not covered the requested mature range
- **THEN** status reads remain based on canonical recurring status
- **AND** availability reads expose incomplete or unavailable finalization evidence rather than reporting compliance
