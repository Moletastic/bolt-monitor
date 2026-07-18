## ADDED Requirements

### Requirement: System associates recurring results with scheduled slots
System SHALL persist the stable scheduled-slot identity and recurring/manual source on execution results used for availability accounting.

#### Scenario: Recurring execution result is persisted
- **WHEN** a recurring execution reaches terminal result processing
- **THEN** its result preserves tenant, service, monitor, schedule-definition version, and UTC scheduled time identity
- **AND** downstream accounting can distinguish it from retries and manual checks
- **AND** the result satisfies an independently derived expected slot rather than creating that expectation

### Requirement: System persists one canonical recurring result per slot
System SHALL establish at most one canonical accounting result for a scheduled slot independently of append-only raw run history.

#### Scenario: Multiple raw runs reference one slot
- **WHEN** retry-safe processing observes duplicate raw execution records for the same scheduled slot
- **THEN** a conditional canonical record identifies one terminal result for objective classification
- **AND** raw history duplication cannot increase availability opportunity counts

#### Scenario: Result has no scheduler work record
- **WHEN** a valid canonical recurring result identifies an independently expected slot but scheduler work evidence is unavailable
- **THEN** the result may satisfy that expected slot under the retry-safe identity contract
- **AND** the absence of work remains available as separate pipeline correlation evidence

#### Scenario: Result arrives after objective correction horizon
- **WHEN** a canonical result is accepted more than 24 hours after its slot matured
- **THEN** raw history preserves the late result and its recurring identity
- **AND** closed objective accounting remains unchanged unless an authorized correction is applied
