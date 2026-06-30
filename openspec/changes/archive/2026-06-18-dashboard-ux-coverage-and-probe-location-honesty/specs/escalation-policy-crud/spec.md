## ADDED Requirements

### Requirement: System excludes service-scoped business hours from escalation policy payloads

The dashboard SHALL NOT include service-scoped business hours in the create or update escalation-policy payload, and SHALL NOT invoke service update APIs from escalation-policy server actions.

#### Scenario: Operator submits new escalation policy

- **WHEN** operator submits the new escalation policy form
- **THEN** the dashboard server action persists the escalation policy through the existing escalation-policy create API
- **AND** the action does not call any service update API as a side effect of policy creation
- **AND** any service-scoped business-hours field present in the submitted form is ignored with a development-mode warning

#### Scenario: Operator submits escalation policy update

- **WHEN** operator submits the escalation policy edit form
- **THEN** the dashboard server action persists the policy through the existing escalation-policy update API
- **AND** the action does not call any service update API as a side effect of policy update

#### Scenario: Service binding is needed later

- **WHEN** a future change introduces an explicit service binding for escalation policies
- **THEN** the binding is exposed as its own API surface and form field, not as a hidden side effect of policy creation

## MODIFIED Requirements

### Requirement: System requires channel references for each step

System SHALL validate that each escalation step specifies a notification channel by `channelId`.

#### Scenario: Operator creates step with no channel

- **WHEN** operator creates an escalation step with no `channelId`
- **THEN** system rejects the request with validation error

#### Scenario: Operator creates step with blank channelId

- **WHEN** operator creates an escalation step whose `channelId` is blank
- **THEN** system rejects the request with validation error pointing at the offending step index

#### Scenario: Dashboard surfaces empty channel option

- **WHEN** operator views the escalation-policy step editor
- **THEN** an empty `channelId` option is presented in the channel select control as `Pick a channel` so that client-side validation matches the server-side rule