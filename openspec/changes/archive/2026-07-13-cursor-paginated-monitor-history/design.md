## Context

Monitor detail currently requests runs, incidents, and audit events on every page render. Runs are limited to 20 but expose no continuation; monitor incidents and monitor/service audit histories use unbounded `Query` calls. Audit reads query the tenant `AUDIT#` range and discard events for other resources in Lambda. DynamoDB caps each query response at 1 MB, so current readers can both consume unnecessary capacity and silently omit older records.

The table already has primary keys that order runs and monitor incidents by RFC3339 timestamp plus stable ID. Audit events have a tenant primary access pattern but no index keyed by resource. Existing `GSI1` is reserved for open incidents, and its schema constant does not match the SST index name, so this change does not reuse it.

## Goals / Non-Goals

**Goals:**
- Read each monitor-history page through one bounded DynamoDB `Query` of at most 20 matching records.
- Provide opaque cursor continuation without counts or offset pagination.
- Read monitor and service audit history through an exact resource-scoped index key.
- Preserve historical audit visibility through an explicit backfill before reader cutover.
- Load monitor-detail evidence tab data only when needed and retain loaded pages while the operator switches tabs.

**Non-Goals:**
- Paginate tenant-wide incidents, incident activities, service cards, or global audit trails.
- Change check-run retention or audit-event business semantics.
- Add infinite scrolling, total-history counts, filters, sorting controls, or direct browser-to-monitor-API requests.

## Decisions

### Add a dedicated sparse audit resource index

Add `GSI3PK` and `GSI3SK` table attributes and an `AuditByResourceIndex`. Every `AuditEventRecord` receives:

```text
GSI3PK = AUDIT_RESOURCE#<tenant>#<service>#<monitor>
GSI3SK = AUDIT#<timestamp>#<auditId>
```

An empty monitor component identifies service-level audit events. The existing primary audit item remains unchanged. Monitor and service audit readers query this index with an exact partition key and `begins_with(GSI3SK, "AUDIT#")`, descending, with `Limit: 20`.

Alternative: reuse `GSI1`. Rejected because it is the open-incident index, its deployed name and schema constant differ, and multi-purpose naming would obscure capacity ownership and future maintenance.

### Use DynamoDB continuation keys as opaque cursors

Each monitor-history repository query accepts an optional cursor and returns records plus `LastEvaluatedKey`. The API serializes this key as base64url JSON and returns it as `nextCursor`; it decodes and validates a supplied cursor before assigning it to `ExclusiveStartKey`.

The cursor is opaque to dashboard callers. It is not an authorization boundary: the endpoint's service and monitor path remains authoritative, and decoded keys are validated against that expected resource key before querying.

Alternative: expose timestamp/ID cursors. Rejected because it couples clients to DynamoDB key encoding and makes later index changes breaking API changes.

### Add cursor pagination alongside existing page pagination

Keep existing page-based `OkPaginated` behavior for unaffected endpoints. Add cursor pagination metadata and a dedicated response constructor for collection endpoints that use it:

```json
"pagination": { "size": 20, "nextCursor": "..." }
```

`nextCursor` is omitted on the final page. Cursor endpoints do not calculate or return a total. Dashboard response types model page and cursor variants without `any`.

Alternative: overload page/total with synthetic values. Rejected because it misrepresents an unknown total and makes continuation semantics ambiguous.

### Keep history sorting inside DynamoDB

Runs, monitor incidents, and indexed audit events already have lexicographically time-ordered sort keys with stable ID tie-breakers. Queries use descending sort order and do not sort or resource-filter results in Go after retrieval.

### Seed runs; lazy-load and cache other evidence tabs

The initial monitor detail server render still retrieves the newest runs page because timeline and metric indicators require it. That page seeds the Runs tab. The evidence-tab client boundary loads Incidents or Audit only on first selection and retains each tab's accumulated rows and cursor in component state. Load-more requests use server actions, append one page, disable while pending, preserve visible rows on failure, and expose retry.

The initial `?tab=` value selects the first visible tab for deep links. Subsequent tab switches are local tab state so loaded tab pages remain cached and do not reissue monitor API requests.

Alternative: navigate with tab links for every selection. Rejected because dashboard API fetches use `no-store`, so navigation recreates the page and loses the cache.

## Risks / Trade-offs

- [Historical audit rows are not indexed automatically] → Backfill all existing `AuditEvent` items and verify count/sample results before reader cutover.
- [GSI reads are eventually consistent] → Audit history is operational context, not a read-after-write control surface; existing primary audit reads do not require strong consistency.
- [Index adds write/storage cost per audit event] → Sparse index is written only for audit events and removes tenant-wide audit reads.
- [New records arrive while paging] → Cursor continuation remains gap-free for records at or older than page boundary; newly created records appear after refresh, not inserted into already loaded pages.
- [Client cache is lost on full navigation] → Cache scope is one mounted monitor-detail page; deep links and refresh load a fresh first page by design.

## Migration Plan

1. Deploy table fields and `AuditByResourceIndex` without changing readers.
2. Deploy audit record writes that populate `GSI3PK` and `GSI3SK`.
3. Run idempotent backfill for existing `AuditEvent` items, restricted to audit records missing either GSI3 key.
4. Verify backfill completion with item counts and monitor/service audit samples.
5. Deploy resource-index audit readers and cursor endpoints.
6. Deploy dashboard lazy evidence tabs and load-more controls.

Rollback reader/UI changes to current primary audit query if index query or backfill validation fails. Keep new index attributes and dual-compatible records; they are additive and safe to retain.

## Open Questions

- Backfill execution environment and progress reporting need selection before implementation: one-off operational command versus deploy-time migration Lambda.
