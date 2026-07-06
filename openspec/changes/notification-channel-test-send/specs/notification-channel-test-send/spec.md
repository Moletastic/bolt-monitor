## ADDED Requirements

### Requirement: System sends real test notifications for registered channels
The system SHALL allow operators to send a real test notification through an existing notification channel without creating an incident, escalation state, route execution, scheduler run, or monitor run.

#### Scenario: Test send succeeds
- **WHEN** an operator requests a test send for an existing notification channel
- **THEN** the system sends a real provider notification to the channel target using the channel's stored type and config
- **AND** the response indicates success with the tested `channelId`
- **AND** no incident, escalation state, route execution, scheduler run, or monitor run is created

#### Scenario: Channel is referenced by a route
- **WHEN** an operator requests a test send for a notification channel that is referenced by one or more notification routes
- **THEN** the system still attempts the test send
- **AND** route references do not block the test action

#### Scenario: Channel does not exist
- **WHEN** an operator requests a test send for an unknown notification channel ID
- **THEN** the system returns a typed not-found error
- **AND** no provider send is attempted

### Requirement: System uses registered channel delivery behavior
The test-send path SHALL use the registered notification sender for the channel type and SHALL merge the channel target into provider config in the same way production escalation delivery does.

#### Scenario: Telegram channel is tested
- **WHEN** a Telegram channel is tested
- **THEN** the system sends through the Telegram sender with the stored bot token and target chat ID

#### Scenario: Email channel is tested
- **WHEN** an email channel is tested
- **THEN** the system sends through the email sender with the stored API key, sender address, and target recipient address

#### Scenario: SMS channel is tested
- **WHEN** an SMS channel is tested
- **THEN** the system sends through the SMS sender with the stored account credentials, sender number, and target destination number

#### Scenario: Webhook channel is tested
- **WHEN** a webhook channel is tested
- **THEN** the system sends through the webhook sender to the stored target URL

#### Scenario: PagerDuty channel is tested
- **WHEN** a PagerDuty channel is tested
- **THEN** the system sends through the PagerDuty sender with the stored routing key or target routing key

### Requirement: System returns actionable sanitized feedback
The test-send endpoint SHALL return typed success or failure feedback. Failure feedback SHALL be actionable but SHALL NOT expose stored credentials, request headers, bot tokens, API keys, auth tokens, account SIDs, or raw provider payloads that may contain secrets.

#### Scenario: Provider rejects test send
- **WHEN** the provider rejects the test notification
- **THEN** the system returns a typed delivery failure error
- **AND** the operator-facing message identifies the channel type and a sanitized failure reason

#### Scenario: Channel configuration is invalid
- **WHEN** the stored channel config cannot be used by the registered sender
- **THEN** the system returns a typed validation or delivery failure error
- **AND** the response identifies the failing channel without exposing secret values

### Requirement: System audits notification channel test sends
The system SHALL record an audit event for every notification channel test-send attempt, including success and failure attempts.

#### Scenario: Successful test send is audited
- **WHEN** a test notification is sent successfully
- **THEN** the system records an audit event containing the channel ID, channel type, success outcome, and timestamp
- **AND** the audit event does not contain channel secret values

#### Scenario: Failed test send is audited
- **WHEN** a test notification fails
- **THEN** the system records an audit event containing the channel ID, channel type, failure outcome, sanitized failure information, and timestamp
- **AND** the audit event does not contain channel secret values

### Requirement: Dashboard exposes channel test-send action
The dashboard SHALL expose a `Send test` action on the notification channel detail page for existing channels.

#### Scenario: Operator sends test from channel detail
- **WHEN** an operator clicks `Send test` on a notification channel detail page
- **THEN** the action shows visible pending feedback while the request is in flight
- **AND** the final success or failure appears inline on the channel detail page
- **AND** the action does not also produce duplicate toast feedback for the same event

#### Scenario: Test send fails from dashboard
- **WHEN** a channel test send fails
- **THEN** the dashboard shows an accessible error message using the typed dashboard error message rules
- **AND** the operator remains on the channel detail page
