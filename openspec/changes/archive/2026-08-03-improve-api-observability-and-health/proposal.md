## Why

API callers cannot correlate an error with Lambda logs, and the public health endpoint is liveness-only without an explicit boundary. Operators need cheap request correlation and a documented split between public liveness and authenticated dependency verification.

## What Changes

- Return generated API request IDs in `X-Request-Id` for monitor API success and error responses.
- Add structured boundary logging with request ID, Lambda invocation ID, route, method, status, and safe operator context.
- Preserve client correlation IDs only as secondary metadata; never let clients replace generated request IDs.
- Document `/api/health` as public liveness.
- Add a low-frequency authenticated synthetic/readiness verification for a protected v1 read path.
- Add idempotent `make setup-readiness` and `make readiness-api` operator commands using target-scoped SSM credentials.
- Add installation, readiness lifecycle, rotation, and teardown guidance for operators.

## Capabilities

### New Capabilities
- `api-request-correlation`: Monitor API responses and boundary logs expose one safe generated request identifier.
- `api-readiness-verification`: Operators can verify authenticated v1 API availability separately from public liveness.

### Modified Capabilities
- `api-health-endpoint`: Health endpoint is explicitly public liveness and remains dependency-independent.
- `api-documentation`: Repository guidance distinguishes liveness from authenticated readiness verification.

## Impact

Affected code includes monitor API response/logging boundaries, API documentation, infrastructure or operations verification scripts, and tests. Logging is metadata-only with existing bounded retention; no high-cardinality metrics or high-frequency probes are introduced.
