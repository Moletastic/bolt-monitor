## MODIFIED Requirements

### Requirement: Notification channel validation emits typed field errors

`validateNotificationChannel` SHALL return `*shared/errors.TypedError{Code: CodeValidationFailed}` with `Details["field"]` set to the offending field's dotted path. Field paths for nested config keys SHALL use the `config.<key>` form.

#### Scenario: Missing required bot token
- **WHEN** a client POSTs a Telegram channel config that omits `config.botToken`
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"config.botToken"`

#### Scenario: Missing required email config keys
- **WHEN** a client POSTs an email channel config that omits `config.apiKey`
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"config.apiKey"`

#### Scenario: Name length validation
- **WHEN** a client POSTs a channel with a name longer than 80 characters
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"name"`
