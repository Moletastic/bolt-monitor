## Why

Incident delivery listing silently drops records after 200 results. Growing collections require bounded reads and an explicit continuation contract so operators can inspect every delivery without unbounded Lambda, DynamoDB, or response cost.

## What Changes

- Replace incident-delivery silent truncation with opaque cursor pagination and a validated bounded `limit`.
- Preserve stable chronological ordering with delivery identity as a tie-breaker across pages.
- Document delivery pagination in OpenAPI and Bruno.
- Add tests for page boundaries, cursor validation, and complete non-duplicated traversal.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `notification-delivery-operations`: Incident delivery listing returns bounded cursor pages rather than silently truncating results.
- `api-response-envelope`: Cursor pagination remains the response-envelope contract for a growing delivery collection.
- `api-documentation`: OpenAPI documents delivery cursor and limit parameters and response continuation metadata.

## Impact

Affected code includes the monitor API delivery handler and repository query, cursor helpers, OpenAPI, Bruno, dashboard delivery reads, and Go/API contract tests. The design avoids total-count queries and limits DynamoDB read work per request.
