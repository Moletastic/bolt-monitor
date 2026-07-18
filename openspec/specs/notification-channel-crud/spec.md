# notification-channel-crud Specification

## Purpose
TBD - created by archiving change notification-channels-and-routes. Update Purpose after archive.
## Requirements
### Requirement: Notification channel registry
The system SHALL store notification channels as first-class entities with `channelId`, `name`, `type`, `target`, `config`, `createdAt`, `updatedAt`.

`name` is the operator-facing label (1-80 chars). `type` is one of: `telegram`, `email`, `sms`, `webhook`, `pagerduty`. `target` is the destination per channel type. `config` is a JSON object holding credentials (bot token, API keys, etc.).

#### Scenario: Create channel
- **WHEN** a `POST /api/v1/notification-channels` request carries a valid `{ name, type, target, config }` body
- **THEN** the system persists the channel and returns `201 Created` with the saved record (including server-assigned `channelId`)

#### Scenario: List channels
- **WHEN** a `GET /api/v1/notification-channels` request arrives
- **THEN** the system returns `200 OK` with `{ channels: [...] }` containing every channel for the tenant

#### Scenario: Get channel by id
- **WHEN** a `GET /api/v1/notification-channels/{channelId}` request arrives for an existing channel
- **THEN** the system returns `200 OK` with the full channel record

#### Scenario: Update channel
- **WHEN** a `PUT /api/v1/notification-channels/{channelId}` request carries a partial `{ name?, type?, target?, config? }` body
- **THEN** the system applies the patch, updates `updatedAt`, and returns `200 OK` with the updated record

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

### Requirement: Channel deletion blocked when referenced
The system SHALL refuse to delete a channel that any notification route references.

#### Scenario: Delete referenced channel
- **WHEN** a `DELETE /api/v1/notification-channels/{channelId}` request arrives AND at least one route step references `channelId`
- **THEN** the system returns `409 Conflict` with `{ error: "channel in use", referencingRoutes: [{ policyId, name }, ...] }`

#### Scenario: Delete unreferenced channel
- **WHEN** a `DELETE /api/v1/notification-channels/{channelId}` request arrives AND no route references `channelId`
- **THEN** the system removes the channel and returns `204 No Content`

### Requirement: Channel credentials redacted in responses
The system SHALL redact secret fields (`botToken`, `apiKey`, `authToken`, `accountSid`) from any channel record returned via the API, replacing each with `***REDACTED***`.

#### Scenario: Get channel returns redacted secrets
- **WHEN** any read endpoint returns a channel record
- **THEN** secret fields appear as `***REDACTED***`; non-secret fields (`name`, `target`, `fromEmail`, `fromNumber`, `routingKey`) appear in cleartext

### Requirement: Dashboard channel CRUD
The dashboard SHALL expose `/integrations/channels` with list, create, edit, and delete views.

#### Scenario: List view
- **WHEN** an operator visits `/integrations/channels`
- **THEN** the dashboard shows a table with columns `Name`, `Type`, `Target`, `Updated`; each row links to the edit view; the page header reads "Notification channels"

#### Scenario: Create view
- **WHEN** an operator visits `/integrations/channels/new`
- **THEN** the dashboard renders a form with `Name`, `Type`, `Target`, and credential inputs appropriate to the selected type; the primary button reads "Create channel"

#### Scenario: Edit view
- **WHEN** an operator visits `/integrations/channels/{channelId}`
- **THEN** the dashboard renders the form pre-populated with the current record; secret fields display as `••••••` masked; the primary button reads "Save changes"; a "Delete channel" button is also present

#### Scenario: Delete confirmation
- **WHEN** an operator clicks "Delete channel"
- **THEN** the dashboard shows a confirm dialog stating how many routes reference the channel before completing the delete

### Requirement: Dashboard channel detail includes test action
The dashboard notification channel detail page SHALL include a non-destructive test-send action for existing channels in addition to edit and delete controls.

#### Scenario: Existing channel detail shows send test action
- **WHEN** an operator views an existing notification channel detail page
- **THEN** the page includes a `Send test` action
- **AND** the action is visually distinct from destructive delete controls

#### Scenario: New channel form does not show send test action
- **WHEN** an operator is creating a new notification channel that has not been saved
- **THEN** the dashboard does not show a `Send test` action

### Requirement: Channel type metadata in dashboard
The dashboard SHALL render the channel type as a human-readable label and show the credential inputs appropriate to the selected type only.

#### Scenario: Telegram credentials
- **WHEN** an operator selects `Type = Telegram` in the create/edit form
- **THEN** the dashboard shows `Bot token` (password input) and `Target = Chat ID`

#### Scenario: Email credentials
- **WHEN** an operator selects `Type = Email`
- **THEN** the dashboard shows `Provider API key` (password), `From address`, optional `API base URL`, and `Target = Recipient email`

#### Scenario: SMS credentials
- **WHEN** an operator selects `Type = SMS (Twilio)`
- **THEN** the dashboard shows `Account SID`, `Auth token` (password), `From number`, optional `API base URL`, and `Target = Destination number`

#### Scenario: Webhook credentials
- **WHEN** an operator selects `Type = Webhook`
- **THEN** the dashboard shows no credential inputs; `Target = Webhook URL`

#### Scenario: PagerDuty credentials
- **WHEN** an operator selects `Type = PagerDuty`
- **THEN** the dashboard shows no extra credential inputs; `Target = Routing key`
