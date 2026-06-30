## ADDED Requirements

### Requirement: System implements 5-state monitor model
System SHALL implement a monitor state machine with exactly 5 states: UP, DEGRADED, DOWN, RECOVERING, and MAINTENANCE.

#### Scenario: Monitor is UP
- **WHEN** monitor has no open incident and consecutive failures are below failure threshold
- **THEN** monitor state is UP

#### Scenario: Monitor is DEGRADED
- **WHEN** monitor has no open incident and consecutive failures are greater than or equal to 1 but less than failure threshold
- **THEN** monitor state is DEGRADED

#### Scenario: Monitor is DOWN
- **WHEN** monitor has an open incident and consecutive failures are greater than or equal to failure threshold
- **THEN** monitor state is DOWN

#### Scenario: Monitor is RECOVERING
- **WHEN** monitor has an open incident and consecutive successes are greater than or equal to 1 but less than recovery threshold
- **THEN** monitor state is RECOVERING

#### Scenario: Monitor is MAINTENANCE
- **WHEN** monitor has been placed in maintenance mode via explicit action
- **THEN** monitor state is MAINTENANCE

### Requirement: System tracks consecutive failure counter
System SHALL track a consecutive failure counter per monitor, persisting it in the MonitorStatus record.

#### Scenario: Failure increments counter
- **WHEN** a check execution results in a failure outcome
- **THEN** system increments ConsecutiveFailures by 1

#### Scenario: Success resets counter
- **WHEN** a check execution results in a success outcome
- **THEN** system resets ConsecutiveFailures to 0

#### Scenario: Manual run does not affect counter
- **WHEN** a check execution is triggered manually via API
- **THEN** system does NOT modify ConsecutiveFailures

#### Scenario: Counter resets on MAINTENANCE entry
- **WHEN** monitor enters MAINTENANCE state
- **THEN** system resets ConsecutiveFailures to 0

### Requirement: System tracks consecutive success counter
System SHALL track a consecutive success counter per monitor, persisting it in the MonitorStatus record.

#### Scenario: Success increments counter
- **WHEN** a check execution results in a success outcome
- **THEN** system increments ConsecutiveSuccesses by 1

#### Scenario: Failure resets counter
- **WHEN** a check execution results in a failure outcome
- **THEN** system resets ConsecutiveSuccesses to 0

#### Scenario: Manual run does not affect counter
- **WHEN** a check execution is triggered manually via API
- **THEN** system does NOT modify ConsecutiveSuccesses

#### Scenario: Counter resets on MAINTENANCE entry
- **WHEN** monitor enters MAINTENANCE state
- **THEN** system resets ConsecutiveSuccesses to 0

### Requirement: System implements state transition logic
System SHALL apply state transitions based on current state, check outcome, and threshold values.

#### Scenario: UP + failure → DEGRADED
- **WHEN** monitor is in UP state and check result is failure
- **THEN** monitor transitions to DEGRADED state with ConsecutiveFailures = 1

#### Scenario: DEGRADED + failure (threshold not met) → DEGRADED
- **WHEN** monitor is in DEGRADED state, check result is failure, and ConsecutiveFailures < FailureThreshold
- **THEN** monitor remains in DEGRADED state with ConsecutiveFailures incremented

#### Scenario: DEGRADED + failure (threshold met) → DOWN
- **WHEN** monitor is in DEGRADED state, check result is failure, and ConsecutiveFailures >= FailureThreshold
- **THEN** monitor transitions to DOWN state and opens an incident

#### Scenario: DEGRADED + success → UP
- **WHEN** monitor is in DEGRADED state and check result is success
- **THEN** monitor transitions to UP state with ConsecutiveFailures reset to 0

#### Scenario: DOWN + failure → DOWN
- **WHEN** monitor is in DOWN state and check result is failure
- **THEN** monitor remains in DOWN state with ConsecutiveFailures incremented and incident still open

#### Scenario: DOWN + success → RECOVERING
- **WHEN** monitor is in DOWN state and check result is success
- **THEN** monitor transitions to RECOVERING state with ConsecutiveSuccesses = 1

#### Scenario: RECOVERING + success (threshold not met) → RECOVERING
- **WHEN** monitor is in RECOVERING state, check result is success, and ConsecutiveSuccesses < RecoveryThreshold
- **THEN** monitor remains in RECOVERING state with ConsecutiveSuccesses incremented

#### Scenario: RECOVERING + success (threshold met) → UP
- **WHEN** monitor is in RECOVERING state, check result is success, and ConsecutiveSuccesses >= RecoveryThreshold
- **THEN** monitor transitions to UP state and resolves the open incident

#### Scenario: RECOVERING + failure → DOWN
- **WHEN** monitor is in RECOVERING state and check result is failure
- **THEN** monitor transitions to DOWN state with ConsecutiveSuccesses reset to 0

#### Scenario: Any state + MAINTENANCE enable → MAINTENANCE
- **WHEN** operator enables maintenance mode on a monitor in any state
- **THEN** monitor transitions to MAINTENANCE state with counters reset to 0

#### Scenario: MAINTENANCE + MAINTENANCE disable → UP
- **WHEN** operator disables maintenance mode on a monitor in MAINTENANCE state
- **THEN** monitor transitions to UP state and re-evaluates current check result

### Requirement: System re-evaluates state on threshold config change
System SHALL re-evaluate monitor state when FailureThreshold or RecoveryThreshold configuration changes while monitor is in DEGRADED or RECOVERING state.

#### Scenario: Threshold reduction triggers immediate re-evaluation
- **WHEN** monitor is in DEGRADED state and FailureThreshold is reduced
- **THEN** if ConsecutiveFailures >= new FailureThreshold, monitor immediately transitions to DOWN and opens incident

#### Scenario: Threshold increase maintains current state
- **WHEN** monitor is in DEGRADED state and FailureThreshold is increased
- **THEN** monitor remains in DEGRADED state and continues accumulating failures under the new threshold
