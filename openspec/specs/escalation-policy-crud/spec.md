# escalation-policy-crud Specification

## Purpose
TBD - created by archiving change notification-channels-and-routes. Update Purpose after archive.
## Requirements
### Requirement: Escalation policy validation emits typed field errors

`validateEscalationPolicy` SHALL return `*shared/errors.TypedError{Code: CodeValidationFailed}` with `Details["field"]` set to the offending field's dotted path. Step-level failures SHALL use indexed bracket paths such as `businessHoursPath.steps[2].channelId`.

#### Scenario: Empty business-hours step channel
- **WHEN** a client POSTs a policy whose business-hours step 3 has an empty `channelId`
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"businessHoursPath.steps[2].channelId"`

#### Scenario: Empty off-hours path
- **WHEN** a client POSTs a policy with no off-hours steps
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"offHoursPath"`

#### Scenario: Inline channel config rejected via INLINE_CHANNEL_CONFIG
- **WHEN** a client POSTs a policy whose step contains `target` or `config` or `channels`
- **THEN** the response body's `reason.code` is `INLINE_CHANNEL_CONFIG` with status 400

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

### Requirement: Dashboard exposes notification route deletion
The dashboard SHALL allow operators to delete an unreferenced notification route from its edit page using the existing escalation-policy delete API.

#### Scenario: Operator deletes an unreferenced notification route
- **WHEN** an operator confirms deletion from a notification route edit page for a route that is not referenced by any service
- **THEN** the dashboard SHALL call the delete escalation-policy action with that route ID
- **AND** the operator SHALL be redirected to the notification routes list with a deletion success message

#### Scenario: Operator attempts to delete a referenced notification route
- **WHEN** an operator confirms deletion from a notification route edit page for a route that is still referenced by a service
- **THEN** the monitor API SHALL reject the delete request with `POLICY_REFERENCED`
- **AND** the dashboard SHALL keep the operator on the route edit page with an inline error message
