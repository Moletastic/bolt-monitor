## Context

`services/monitor-api` already handles `GET /api/v1/probe-locations`, but `infra/stacks/bootstrap.ts` does not register that route on the bootstrap API Gateway. Result: source code and OpenSpec claim route exists, while live SST-managed API surface does not expose it.

## Decision

- Add one API Gateway route: `GET /api/v1/probe-locations` using the same `monitorHandler` used by other monitor routes.

## Non-Goals

- No dashboard form rewrite.
- No probe-location catalog model changes.
- No auth, tenancy, or entitlement work.
