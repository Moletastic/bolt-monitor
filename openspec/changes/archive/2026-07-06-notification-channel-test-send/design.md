## Context

Notification channels are already first-class records under `notification-channel-crud`, and escalation dispatch already resolves a route step's `channelId` into a stored channel before sending through `shared/notifications` senders. The dashboard can create, edit, and guard deletion of channels, but there is no operator-facing way to verify that a saved destination actually receives messages.

The useful test is not a dry-run validation. Operators need to know whether the registered bot token, webhook URL, routing key, API key, or target address works against the real provider. This design therefore treats test-send as a real delivery action with harmless content and visible audit history.

## Goals / Non-Goals

**Goals:**
- Add a real test-send action for existing notification channels.
- Reuse the same provider senders used by escalation dispatch so test behavior matches production delivery as closely as practical.
- Support Telegram, email, SMS, webhook, and PagerDuty channels.
- Allow test sends for channels currently referenced by notification routes.
- Return typed, actionable success or failure feedback to the dashboard without exposing secrets.
- Record an audit event for each test-send attempt.

**Non-Goals:**
- Do not create incidents, escalation states, route executions, or scheduler runs.
- Do not test notification route ordering, delays, or service binding behavior.
- Do not introduce new external dependencies or a background worker for this path.
- Do not expose stored credentials, raw tokens, or full provider payloads in API responses or audit records.
- Do not require the channel to be unused by routes.

## Decisions

### Use real provider delivery

Decision: `POST /api/v1/notification-channels/{channelId}/test` will send a real notification to the configured destination.

Rationale: A validation-only endpoint would confirm shape, not deliverability. The operator question is "will this alert arrive?", so the endpoint should exercise the actual provider integration.

Alternative considered: Config-only validation. Rejected because it would miss invalid Telegram chat IDs, revoked provider keys, blocked webhook endpoints, and other real delivery failures.

### Reuse the notification sender path

Decision: The monitor API should build a `notifications.Notification` with test content and send it through a `notifications.SenderRegistry` using the same channel type senders as escalation runtime.

Rationale: Reusing `shared/notifications` keeps test behavior consistent with escalation delivery and avoids duplicating provider-specific code in dashboard or monitor API layers.

Alternative considered: Call provider APIs directly from monitor API handlers. Rejected because it duplicates provider logic and creates drift from production delivery.

### Keep test-send channel-scoped, not route-scoped

Decision: The endpoint tests a single channel record and does not create a route execution or incident activity.

Rationale: The capability is about registered destination correctness. Route ordering and escalation behavior should remain covered by escalation-runtime flows and separate tests.

Alternative considered: Add a "test route" button. Rejected for this change because it would involve delays, route state, service binding, and incident semantics.

### Audit every attempt

Decision: Each test-send attempt records an audit event with channel ID, channel type, success/failure, and sanitized failure category or message.

Rationale: Real sends are operator actions with possible external effects. Auditability helps explain why a test message appeared and who initiated it once actor context exists.

Alternative considered: Audit only successful sends. Rejected because failures are often the most operationally useful events to inspect.

### Sanitize provider errors

Decision: API errors may include provider status/category and a concise sanitized message, but MUST NOT include secrets, request headers, bot tokens, API keys, auth tokens, account SIDs, or raw provider payloads that may contain credentials.

Rationale: Operators need actionable diagnostics, but notification credentials are sensitive.

Alternative considered: Return raw provider errors. Rejected because provider errors can echo request details or identifiers not intended for broad dashboard display.

## Risks / Trade-offs

- [Risk] Test sends can notify real people or external systems. -> Mitigation: Use explicit button copy, pending feedback, and a harmless test message stating no incident was created.
- [Risk] Provider errors may leak secrets. -> Mitigation: sanitize errors before response/audit storage and add tests for redaction boundaries.
- [Risk] Reusing escalation sender setup inside monitor API may duplicate registry construction. -> Mitigation: keep registry construction small and shared where practical without creating a new service abstraction unless needed.
- [Risk] PagerDuty test sends may create real events. -> Mitigation: use a clear test summary/source and document that this is a real send; future work can add per-type preview warnings if needed.
- [Risk] Audit storage model may not have a channel-specific event shape. -> Mitigation: reuse existing audit event conventions where available and keep details generic.

## Migration Plan

1. Add typed API behavior and sender wiring behind the new endpoint.
2. Add audit recording for success and failure attempts.
3. Add dashboard server action and channel detail button with typed action-state feedback.
4. Verify each channel type through unit tests with fake senders; manually verify at least one real configured channel in staging.
5. Rollback is safe by removing the dashboard button and endpoint route; existing channel records and routes are unchanged.

## Open Questions

- What exact audit event type naming convention should be used for channel test sends?
- Should PagerDuty test messages use `incident.down` semantics or a distinct test event payload if the sender requires a provider-specific event action?
- Is actor identity currently available in dashboard server actions, or should audit records use the existing default/system actor convention?
