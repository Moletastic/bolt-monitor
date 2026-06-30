## MODIFIED Requirements

### Requirement: INLINE_CHANNEL_CONFIG is a typed code constant

The `INLINE_CHANNEL_CONFIG` code SHALL be a typed constant in `shared/errors`. The handler that enforces the "steps must reference channels by channelId" rule SHALL emit `*shared/errors.TypedError{Code: CodeInlineChannelConfig}` rather than building the error envelope inline.

#### Scenario: Step with inline target rejected
- **WHEN** a client POSTs an escalation policy whose step contains an inline `target`
- **THEN** the response body's `reason.code` is `INLINE_CHANNEL_CONFIG` and the status is 400
