## 1. API Contract And Error Model

- [x] 1.1 Add or reuse typed error codes for notification test delivery failure and any required validation/not-found cases.
- [x] 1.2 Add `POST /api/v1/notification-channels/{channelId}/test` routing in the monitor API without changing existing channel CRUD routes.
- [x] 1.3 Define the success response payload for channel test sends, including `channelId` and a server-generated timestamp.
- [x] 1.4 Ensure API responses use the standard response envelope and route errors through shared error handling.

## 2. Test-Send Delivery Path

- [x] 2.1 Add sender registry construction for monitor API test sends using existing `shared/notifications` senders.
- [x] 2.2 Load the stored channel by ID and return a typed not-found error when it is missing.
- [x] 2.3 Merge the channel target into provider config using behavior equivalent to escalation dispatch.
- [x] 2.4 Build a harmless test notification message that states no incident was created.
- [x] 2.5 Send through the registered sender for Telegram, email, SMS, webhook, and PagerDuty channel types.
- [x] 2.6 Sanitize provider/config failure details before returning errors or storing audit details.

## 3. Audit Events

- [x] 3.1 Identify the existing audit event writer/model that best fits notification channel actions.
- [x] 3.2 Record a success audit event for each successful test send with channel ID, type, outcome, and timestamp.
- [x] 3.3 Record a failure audit event for each failed test send with channel ID, type, sanitized failure information, outcome, and timestamp.
- [x] 3.4 Add tests that audit records never include bot tokens, API keys, auth tokens, account SIDs, or raw secret config.

## 4. Dashboard Action

- [x] 4.1 Add a dashboard API helper for `POST /api/v1/notification-channels/{channelId}/test`.
- [x] 4.2 Add a server action returning typed `ActionState` for channel test sends.
- [x] 4.3 Add a `Send test` button to existing channel detail pages only, visually separated from delete controls.
- [x] 4.4 Show pending, success, and error feedback inline on the channel detail page without duplicate toast feedback.
- [x] 4.5 Keep the operator on the channel detail page after success or failure.

## 5. Coverage

- [x] 5.1 Add monitor API tests for successful test sends using fake senders.
- [x] 5.2 Add monitor API tests for unknown channel IDs, unsupported sender types, invalid config, and provider failure.
- [x] 5.3 Add monitor API tests proving channels referenced by routes can still be tested.
- [x] 5.4 Add dashboard tests or guard coverage for the channel detail `Send test` action and typed feedback surfaces.
- [x] 5.5 Add sender/config tests for sanitized delivery failure output where practical.

## 6. Verification

- [x] 6.1 Run `make test-go-all`.
- [x] 6.2 Run `make lint-go`.
- [x] 6.3 Run `make lint-dashboard`.
- [x] 6.4 Run `make check-dashboard`.
- [x] 6.5 Run `make test-dashboard`.
- [ ] 6.6 Manually test at least one real configured channel in staging and confirm the dashboard shows a single inline result.
