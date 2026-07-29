## ADDED Requirements

### Requirement: Scheduled-step queue envelopes are consumed canonically
The notification SQS consumer SHALL validate canonical `scheduled_step` envelopes and invoke scheduled escalation handling using their incident, tenant, transition, and step identity. It SHALL report malformed or unsupported scheduled envelopes as partial SQS failures.

#### Scenario: Scheduler sends a delayed step
- **WHEN** EventBridge Scheduler delivers a valid canonical scheduled-step message to notification SQS
- **THEN** the runtime executes the identified delayed escalation step through the queue path
- **AND** does not reinterpret it as a legacy notification event

#### Scenario: Scheduled message is malformed
- **WHEN** a scheduled-step message lacks valid required identity or step number
- **THEN** the runtime reports only that SQS message as failed
- **AND** notification queue redrive handles it through existing DLQ policy
