## Why

Monitor detail eagerly reads runs, incidents, and audit history even when an operator views only one tab. Incident and audit reads can grow without a bound, while monitor audit history reads the tenant-wide audit range and filters unrelated records in Lambda. This increases DynamoDB read cost, Lambda work, and response latency as history grows.

## What Changes

- Add cursor-based, newest-first pagination to monitor runs, incidents, and audit history, with a fixed page size of 20 and opaque continuation cursors.
- Add a resource-scoped DynamoDB audit index so monitor and service audit history reads do not query and filter tenant-wide audit events.
- Backfill existing audit events before moving audit reads to the new index so historical audit visibility is preserved.
- Extend the shared response envelope with cursor pagination metadata and add matching dashboard types and parsing.
- Change monitor detail evidence tabs to lazy-load and retain each tab's first page in client memory; add an outlined `Load more` button that appends each following page.

## Capabilities

### New Capabilities
- `monitor-history-cursor-pagination`: Cursor-based retrieval of monitor runs, incidents, and audit history.

### Modified Capabilities
- `api-response-envelope`: Paginated collection envelopes support opaque cursor continuation metadata.
- `audit-event-read-api`: Monitor and service audit history reads are bounded and resource-scoped.
- `dynamodb-single-table-storage`: Audit events have a documented resource-scoped secondary-index access pattern.
- `dashboard-web-app`: Monitor evidence tabs load history on demand, retain loaded tab data, and let operators append older records.

## Impact

- Affects `infra/stacks/bootstrap.ts`, DynamoDB schema and record constructors, audit migration tooling, monitor-api repositories and handlers, shared response envelopes, dashboard API types, server actions, and monitor detail UI.
- Changes pagination metadata for newly paginated monitor history endpoints; consumers must use `nextCursor` rather than page totals.
- Adds one sparse audit secondary index and indexed attributes to audit-event writes.
