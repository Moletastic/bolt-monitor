## ADDED Requirements

### Requirement: Incident operator commands are idempotent
Incident acknowledgement and resolution SHALL require `Idempotency-Key` and return one bounded canonical result for each scoped key and request.

#### Scenario: Incident command is retried
- **WHEN** an operator retries acknowledgement or resolution with the same key and request
- **THEN** the system returns the original incident result
- **AND** records at most one command mutation

#### Scenario: Concurrent incident commands use the same key
- **WHEN** equivalent acknowledgement or resolution requests race with the same scoped key
- **THEN** one DynamoDB transaction reserves the command record and applies the incident mutation
- **AND** both requests converge on the same stored incident result
