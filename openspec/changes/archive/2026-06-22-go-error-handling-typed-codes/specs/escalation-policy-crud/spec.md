## MODIFIED Requirements

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
