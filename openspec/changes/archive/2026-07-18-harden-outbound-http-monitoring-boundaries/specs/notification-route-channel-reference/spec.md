## MODIFIED Requirements

### Requirement: Channel resolution at dispatch
When the escalation runtime fires a step, the system SHALL resolve the referenced channel and dispatch using the channel's `target` and `config`. Every HTTP-based sender request SHALL use the shared default public outbound policy and SHALL return sanitized typed delivery failures to the runtime.

#### Scenario: Resolve channel for step dispatch
- **WHEN** escalation runtime processes a step with `channelId`
- **THEN** the system loads the channel record, merges `target` into the appropriate per-type config field (`chatId` for telegram, `toEmail` for email, `toNumber` for sms, `url` for webhook, `routingKey` for pagerduty), and passes the merged config to the registered sender
- **AND** the sender enforces the shared public outbound policy for each HTTP request

#### Scenario: Channel deleted mid-escalation
- **WHEN** escalation runtime attempts to resolve a `channelId` that was deleted between schedule and dispatch
- **THEN** the system logs the missing channel and stops the escalation for that step without retry

#### Scenario: Persisted channel resolves to blocked address
- **WHEN** a resolved channel's webhook or provider endpoint is blocked at dispatch time
- **THEN** the sender makes no request to that endpoint and returns a typed delivery failure without exposing channel secrets
