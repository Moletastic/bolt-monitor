## Why

The deployed API can return an HTTP conflict inside a success envelope, while OpenAPI describes error responses as successes. Consumers need one machine-readable contract for runtime behavior, documentation, and tests.

## What Changes

- Return channel-in-use deletion conflicts through the standard error envelope with a stable typed code and safe reference details.
- Make OpenAPI define success and error envelopes separately and attach request, path, query, pagination, location, and idempotency metadata to operations.
- Extend deterministic API contract checks to reject known semantic mismatches between runtime conventions and OpenAPI.
- Add representative handler contract tests for success, client failure, authorization failure, not found, conflict, and deletion behavior.

## Capabilities

### New Capabilities
- `api-contract-semantic-validation`: Repository release checks detect success/error envelope and operation-metadata contract regressions.

### Modified Capabilities
- `api-response-envelope`: Failure HTTP responses always use the error envelope and stable typed reason code.
- `notification-channel-crud`: Deleting a channel referenced by an escalation policy returns a typed conflict response.
- `api-documentation`: OpenAPI documents operation request, response, error, pagination, location, and idempotency semantics.
- `api-contract-release-gates`: API contract gates validate semantic response and operation metadata requirements.

## Impact

Affected code includes `services/monitor-api`, `shared/errors`, `openapi/openapi.yaml`, API contract scripts, Bruno requests, and Go/Node contract tests. This change does not alter resource routes.
