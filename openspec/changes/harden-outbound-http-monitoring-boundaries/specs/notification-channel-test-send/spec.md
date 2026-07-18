## MODIFIED Requirements

### Requirement: System uses registered channel delivery behavior
The test-send path SHALL use the registered notification sender for the channel type, SHALL merge the channel target into provider config in the same way production escalation delivery does, and SHALL execute every resulting HTTP request through the shared default public outbound policy.

#### Scenario: Telegram channel is tested
- **WHEN** a Telegram channel is tested
- **THEN** the system sends through the Telegram sender with the stored bot token and target chat ID using bounded public outbound HTTP behavior

#### Scenario: Email channel is tested
- **WHEN** an email channel is tested
- **THEN** the system sends through the email sender with the stored API key, sender address, target recipient address, and policy-validated provider endpoint

#### Scenario: SMS channel is tested
- **WHEN** an SMS channel is tested
- **THEN** the system sends through the SMS sender with the stored account credentials, sender number, target destination number, and policy-validated provider endpoint

#### Scenario: Webhook channel is tested
- **WHEN** a webhook channel is tested
- **THEN** the system sends through the webhook sender only when the stored target and every redirect remain permitted by the public outbound policy

#### Scenario: PagerDuty channel is tested
- **WHEN** a PagerDuty channel is tested
- **THEN** the system sends through the PagerDuty sender with the stored routing key and policy-validated provider endpoint

### Requirement: System returns actionable sanitized feedback
The test-send endpoint SHALL return typed success or failure feedback. Policy rejection and transport failures SHALL map to the existing typed validation or notification-delivery response using stable failure categories. Failure feedback SHALL be actionable but SHALL NOT expose stored credentials, URL user information or sensitive query values, request or response headers, bot tokens, API keys, auth tokens, account SIDs, request bodies, or raw provider response payloads.

#### Scenario: Provider rejects test send
- **WHEN** the provider rejects the test notification
- **THEN** the system returns a typed delivery failure error
- **AND** the operator-facing message identifies the channel type and a sanitized failure reason

#### Scenario: Channel configuration is invalid
- **WHEN** the stored channel config cannot be used by the registered sender
- **THEN** the system returns a typed validation or delivery failure error
- **AND** the response identifies the failing channel without exposing secret values

#### Scenario: Persisted destination is now blocked
- **WHEN** a stored webhook target or provider endpoint resolves to a blocked address during test send
- **THEN** the system attempts no provider request and returns `NOTIFICATION_DELIVERY_FAILED` with a stable sanitized blocked-destination reason

#### Scenario: Redirect attempts credential exfiltration
- **WHEN** a provider or webhook response redirects a credentialed request to another origin or from HTTPS to HTTP
- **THEN** the system rejects the redirect rather than following it with the body, configured headers, or credentials stripped
- **AND** it returns a typed sanitized delivery failure

#### Scenario: Provider returns a large error body
- **WHEN** a provider response body exceeds 1 MiB or contains secrets
- **THEN** the system stops reading at the limit and neither API feedback nor audit details contain the raw body
