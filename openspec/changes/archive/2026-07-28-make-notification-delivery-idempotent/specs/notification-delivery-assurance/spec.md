## ADDED Requirements

### Requirement: Provider sends require durable delivery claim
Before each notification provider request, the runtime SHALL create or reuse deterministic delivery records, conditionally claim each eligible delivery, and persist a terminal or retryable outcome with its fencing token. A worker without a claim SHALL NOT call a provider.

#### Scenario: Duplicate SQS transition is delivered
- **WHEN** duplicate transition messages reach concurrent workers
- **THEN** at most one worker claims each pending channel delivery
- **AND** non-claiming workers do not send that channel

#### Scenario: Provider accepts then later work fails
- **WHEN** provider acceptance is followed by an error while scheduling or persisting later escalation work
- **THEN** the accepted channel is persisted or recovered as delivered or ambiguous
- **AND** retry processing does not blindly resend a known delivered channel
