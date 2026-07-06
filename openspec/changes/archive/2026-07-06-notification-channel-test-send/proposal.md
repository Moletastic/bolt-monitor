## Why

Operators can create notification channels, but they cannot verify that a registered channel actually delivers messages until a real incident or escalation fires. A simple test-send action gives immediate confidence that credentials, targets, and provider connectivity are correct before the channel is used in production routes.

## What Changes

- Add a real test-send action for registered notification channels.
- Add `POST /api/v1/notification-channels/{channelId}/test` to load the stored channel, send a harmless test notification through the existing notification sender path, and return typed success or failure feedback.
- Record an audit event for each test-send attempt, including success or failure metadata without exposing secrets.
- Add a dashboard `Send test` button on the notification channel detail page with visible pending, success, and error feedback.
- Support all registered channel types: Telegram, email, SMS, webhook, and PagerDuty.
- Preserve existing route/channel semantics; channels may be tested even when referenced by notification routes.

## Capabilities

### New Capabilities
- `notification-channel-test-send`: Covers real test delivery for registered notification channels, API behavior, auditability, dashboard feedback, and sanitized provider errors.

### Modified Capabilities
- `notification-channel-crud`: Extends channel detail behavior with a non-mutating test-send action for existing channels.

## Impact

- Affected API area: `services/monitor-api` notification channel routes and response/error handling.
- Affected shared code: notification sender reuse and possibly shared error code registration.
- Affected storage/audit area: notification channel test-send audit events.
- Affected dashboard area: notification channel detail page, server actions, typed action-state feedback, and tests.
- No new external dependencies are expected.
- The action sends real provider requests and may produce real messages in configured destinations.
