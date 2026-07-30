## Why

Retries after a lost API response can duplicate operational effects or return a different result for the same manual-run request. The retry contract must converge on one durable outcome before more operator commands rely on it.

## What Changes

- Make delivery replay read `Idempotency-Key` case-insensitively, matching API Gateway HTTP API header behavior.
- Make manual monitor run return one canonical in-progress or completed result for repeat requests using the same scoped key and fingerprint.
- Define bounded idempotency behavior for selected side-effecting operator commands, beginning with notification test sends and incident acknowledgement and resolution.
- Return typed idempotency conflicts when a key is reused for a different request.

## Capabilities

### New Capabilities
- `operator-command-idempotency`: Bounded idempotent retry behavior for side-effecting operator commands beyond manual runs and delivery replay.

### Modified Capabilities
- `manual-run-api`: Same-key retries return the canonical in-progress or completed manual-run result.
- `notification-delivery-operations`: Delivery replay accepts the required idempotency header independent of header casing.
- `notification-channel-test-send`: Notification test sends use bounded idempotency to prevent duplicate external notifications.
- `incident-management-api`: Incident acknowledgement and resolution use bounded idempotency to prevent duplicate command effects.

## Impact

Affected code includes `services/monitor-api` command handlers, idempotency persistence, typed errors, OpenAPI and Bruno definitions, and Go contract tests. DynamoDB idempotency records remain TTL-bounded to control storage cost.
