## ADDED Requirements

### Requirement: System requires channel references for each step
System SHALL validate that each escalation step specifies a notification channel by `channelId`.

#### Scenario: Operator creates step with no channel
- **WHEN** operator creates an escalation step with no `channelId`
- **THEN** system rejects the request with validation error

#### Scenario: Operator creates step with blank channelId
- **WHEN** operator creates an escalation step whose `channelId` is blank
- **THEN** system rejects the request with validation error pointing at the offending step index
