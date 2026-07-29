## Context

`GET /api/v1/incidents/{incidentId}/deliveries` is declared but generic incident detail dispatch matches first. Focused handler tests bypass route dispatch.

## Goals / Non-Goals

**Goals:** Make documented route reach delivery query. Add realistic API Gateway dispatch test.

**Non-Goals:** Change endpoint path, response schema, storage query, pagination, or cache behavior.

## Decisions

- Match specific incident subresources before generic incident detail. Specific-first routing is minimal and keeps existing API contract.
- Test `handleRequest` with actual raw path/method, not only direct handler calls.
- No infrastructure cost change. Fix may eliminate wasteful client retries without cache or new observability resources.

## Risks / Trade-offs

- Future subresources can regress -> Route-level tests guard declared subresource routes.

## Migration Plan

Deploy API binary. No data migration or compatibility window needed; route already documented.
