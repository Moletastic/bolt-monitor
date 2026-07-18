## MODIFIED Requirements

### Requirement: Notification channel validation emits typed field errors
`validateNotificationChannel` SHALL return `*shared/errors.TypedError{Code: CodeValidationFailed}` with `Details["field"]` set to the offending field's dotted path. Field paths for nested config keys SHALL use the `config.<key>` form. Webhook targets and configurable provider API base URLs SHALL be absolute HTTP or HTTPS URLs accepted by the shared default public outbound policy, and validation details SHALL be sanitized.

#### Scenario: Missing required bot token
- **WHEN** a client POSTs a Telegram channel config that omits `config.botToken`
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"config.botToken"`

#### Scenario: Missing required email config keys
- **WHEN** a client POSTs an email channel config that omits `config.apiKey`
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"config.apiKey"`

#### Scenario: Name length validation
- **WHEN** a client POSTs a channel with a name longer than 80 characters
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"name"`

#### Scenario: Webhook target is unsafe
- **WHEN** a client creates or updates a webhook channel whose `target` is non-HTTP, malformed, blocked, or currently resolves to a blocked address
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"target"`
- **AND** no channel mutation is persisted

#### Scenario: Provider base URL is unsafe
- **WHEN** a client creates or updates an email, SMS, or PagerDuty channel whose `config.apiBaseUrl` is non-HTTP, malformed, blocked, or currently resolves to a blocked address
- **THEN** the response body's `reason.code` is `VALIDATION_FAILED` and `reason.details.field` is `"config.apiBaseUrl"`
- **AND** no channel mutation is persisted

#### Scenario: Safe public channel destination remains valid
- **WHEN** a webhook target or provider base URL is a permitted public HTTP or HTTPS URL
- **THEN** URL policy validation does not prevent channel creation or update
