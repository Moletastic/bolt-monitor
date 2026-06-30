## ADDED Requirements

### Requirement: System starts escalation when monitor transitions to DOWN
System SHALL initiate escalation path execution when a monitor transitions to DOWN state and the associated service has an escalation policy assigned.

#### Scenario: Monitor goes DOWN with service having escalation policy
- **WHEN** monitor transitions to DOWN state
- **AND** the parent service has an escalation policy assigned
- **THEN** system creates an escalation state record for the incident
- **AND** system evaluates current time against the service's business hours configuration
- **AND** system selects the appropriate escalation path (business-hours or off-hours)
- **AND** system fires step 1 immediately (delayMinutes = 0)

#### Scenario: Monitor goes DOWN with service having no escalation policy
- **WHEN** monitor transitions to DOWN state
- **AND** the parent service has no escalation policy assigned
- **THEN** system does not create escalation state
- **AND** no notification is sent

### Requirement: System executes steps sequentially with delays
System SHALL fire escalation steps in order, waiting the specified delayMinutes between each step.

#### Scenario: Step 1 has 0 delay
- **WHEN** step 1 has delayMinutes = 0
- **THEN** system fires the step immediately upon escalation start

#### Scenario: Step 2 has non-zero delay
- **WHEN** step 1 has fired and step 2 has delayMinutes = 15
- **THEN** system schedules step 2 to fire 15 minutes after step 1 fires
- **AND** system uses CloudWatch scheduled rule to invoke escalation handler at the scheduled time

#### Scenario: CloudWatch rule fires after Lambda restart
- **WHEN** a CloudWatch scheduled rule triggers an escalation handler invocation
- **AND** the associated incident has already resolved
- **THEN** system skips firing the step and suppresses remaining steps

### Requirement: System persists escalation state in DynamoDB
System SHALL store escalation state per incident, tracking which steps have fired and which is next.

#### Scenario: Escalation state is created
- **WHEN** escalation starts for an incident
- **THEN** system creates an ESCALATION_STATE record containing incidentId, policyId, currentStep, stepsFired list, and status ACTIVE

#### Scenario: Step fires and state advances
- **WHEN** an escalation step fires
- **THEN** system updates the ESCALATION_STATE record: add step to stepsFired, increment currentStep

#### Scenario: Incident resolves before all steps fire
- **WHEN** an incident resolves (monitor goes UP) while escalation is ACTIVE
- **THEN** system updates ESCALATION_STATE status to SUPPRESSED
- **AND** no further steps fire

### Requirement: System creates EscalationExhausted incident when path completes
System SHALL create an escalation.exhausted incident when all escalation steps have fired and the original incident remains open.

#### Scenario: All steps exhausted, incident still open
- **WHEN** the final escalation step fires
- **AND** the original incident is still in OPEN status
- **THEN** system creates a new incident with type escalation.exhausted
- **AND** the new incident contains a reference to the original incidentId
- **AND** system emits escalation.exhausted notification event
- **AND** the original incident remains OPEN

#### Scenario: All steps exhausted, incident already resolved
- **WHEN** the final escalation step fires
- **AND** the original incident is already resolved
- **THEN** system does NOT create an escalation.exhausted incident

### Requirement: System supports multiple channel types
System SHALL support firing notification steps to multiple channel types: telegram, email, sms, webhook, pagerduty.

#### Scenario: Step fires telegram channel
- **WHEN** an escalation step specifies a channel of type telegram
- **THEN** system sends a Telegram message to the configured target

#### Scenario: Step fires email channel
- **WHEN** an escalation step specifies a channel of type email
- **THEN** system sends an email to the configured email address

#### Scenario: Step fires sms channel
- **WHEN** an escalation step specifies a channel of type sms
- **THEN** system sends an SMS to the configured phone number

#### Scenario: Step fires webhook channel
- **WHEN** an escalation step specifies a channel of type webhook
- **THEN** system POSTs to the configured webhook URL with incident payload

#### Scenario: Step fires pagerduty channel
- **WHEN** an escalation step specifies a channel of type pagerduty
- **THEN** system creates a PagerDuty incident via their API integration
