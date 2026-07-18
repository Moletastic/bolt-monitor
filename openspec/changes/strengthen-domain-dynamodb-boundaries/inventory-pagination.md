# §1.4 Inventory: Pagination Behavior on DynamoDB Reads

This inventory captures the pre-change state of every primary-index read
method in the three runtimes. It distinguishes:

- **Exposed** — the helper returns a page with `NextKey` so the caller can choose
  bounded traversal or incomplete-result reporting.
- **Truncated** — the helper returns a single page, silently discarding
  `LastEvaluatedKey`. Callers cannot detect that the response was incomplete.
- **Unbounded** — the helper iterates the full partition in one call by ignoring
  pagination entirely.

All call sites use `sharedaws.DynamoDBAPI` except escalation-runtime, which
still imports `bolt-monitor/shared/dynamodb` (the secondary facade). The §3.4
facade migration moves escalation-runtime onto `sharedaws.DynamoDBAPI`.

## `services/monitor-api`

| Method | Source | Pagination behavior | Plan |
| --- | --- | --- | --- |
| `GetService` | `repository.go:159` | Exact-key `GetItem` (n/a) | — |
| `GetServiceStatus` | `repository.go:278` | Exact-key `GetItem` (n/a) | — |
| `ListServices` | `repository.go:96` | Calls `queryPartition` → `client.Query` **without** `ExclusiveStartKey`; silently iterates items | §4.4 — migrate to page helper that returns `NextKey`; caller chooses bounded behavior |
| `ListMonitors` | `repository.go:347` | Calls `queryPartition` → no cursor | §4.4 — migrate to page helper |
| `serviceMonitorSummaries` | `repository.go:141` | Wraps `ListMonitors` (transitively truncated) | §4.4 — follow `ListMonitors` migration |
| `ListMonitorRuns` | `repository.go:683` | Wraps `ListMonitorRunsPage` (already exposes `NextKey`) | No change to behavior; covered by regression tests in §6.1 |
| `ListMonitorRunsPage` | `repository.go:691` | Exposes `NextKey` directly | Promote to page helper in §4.4; add cursor round-trip test in §4.3 |
| `ListIncidents` | `repository.go:864` | `queryPartition` without cursor | §4.4 — migrate to page helper |
| `ListMonitorIncidents` | `repository.go:941` | Wraps `ListMonitorIncidentsPage` (already exposes `NextKey`) | Regression-test in §6.1 |
| `ListMonitorIncidentsPage` | `repository.go:949` | Exposes `NextKey` directly | Promote to page helper in §4.4 |
| `ListServiceIncidents` | `repository.go:981` | Direct `client.Query` with `Limit` and fixed prefix; **drops** `LastEvaluatedKey` | §4.4 — must expose continuation; service detail UI cannot silently truncate |
| `ListIncidentActivities` | `repository.go:918` | `queryPartition` without cursor | §4.4 — migrate to page helper |
| `ListMonitorAuditEvents` | `repository.go:1090` | Wraps `ListMonitorAuditEventsPage` (GSI query) | Regression-test in §6.1 |
| `ListMonitorAuditEventsPage` | `repository.go:1098` | GSI query, exposes `NextKey` | Keep as named method per §4.6 (GSI semantics) |
| `ListServiceAuditEvents` | `repository.go:1137` | Wraps `ListServiceAuditEventsPage` (GSI query) | Regression-test in §6.1 |
| `ListServiceAuditEventsPage` | `repository.go:1145` | GSI query, exposes `NextKey` | Keep as named method per §4.6 |
| `collectMonitorDeleteKeys` | `repository.go:1184` | `queryPartition` without cursor; collects all items in partition for deletion | Acceptable for delete operations; keep as-is in §4.4 (in-process fan-out) |
| `buildServiceStatusRecord` | `repository.go:1207` | Wraps `ListMonitors` (transitively truncated) | §4.4 — follow `ListMonitors` migration |
| `GetServiceCardMetrics` | `repository.go:743` | Uses `ListMonitorRuns(limit=20)` (single page; bounded by 20) | §6.2 — assert the bounded choice is intentional; failure means the dashboard must report incompleteness |
| `Search` (`search.go`) | `search.go` (whole file) | GSI scan, paginates internally, returns aggregated results | §4.6 — keep GSI-named method; assert `Limit` is explicit |

## `services/check-runtime`

| Method | Source | Pagination behavior | Plan |
| --- | --- | --- | --- |
| `GetMonitor` | `repository.go:40` | Exact-key `GetItem` (n/a) | — |
| `GetLastExecution` | `repository.go:103` | Wraps `getMonitorRecord` (n/a) | — |
| `ListMonitors` | `repository.go:54` | Iterates services then per-service `MONITOR#` query, **never** pages; unbounded by design | §4.5 — bounded by `Limit` per inner call; expose explicit page helper; continue-with-cursor helper must be available for scheduler fan-out |
| `EnqueueExecutionRequests` | `repository.go:138` | Writes only | — |
| `getMonitorRecord` | `repository.go:33` | Exact-key `GetItem` (n/a) | — |
| `GetExecutionWork` | `repository.go:163` (referenced) | Single `GetItem` on `TENANT#<id>` + `RUN_REQUEST#…` | — |
| `ClaimExecutionWork` / scan flow | `repository.go:560` | Per-partition Query without cursor | §4.5 — bounded by explicit `Limit` and `ExclusiveStartKey` propagation |

## `services/escalation-runtime`

| Method | Source | Pagination behavior | Plan |
| --- | --- | --- | --- |
| `GetService` | `repository.go:29` | Exact-key `GetItem` (n/a) | §3.4 — switch to `sharedaws.DynamoDBAPI` |
| `GetEscalationPolicy` | `repository.go:47` | Exact-key `GetItem` (n/a) | §3.4 — switch to `sharedaws.DynamoDBAPI` |
| `GetChannel` | `repository.go:66` | Exact-key `GetItem` (n/a) | §3.4 — switch to `sharedaws.DynamoDBAPI` |
| `PutEscalationState` | `repository.go:85` | `PutItem` (n/a) | §3.4 — switch to `sharedaws.DynamoDBAPI` |
| `GetEscalationState` | `repository.go:94` | Exact-key `GetItem` (n/a) | §3.4 — switch to `sharedaws.DynamoDBAPI` |
| `GetIncident` | `repository.go:116` | Exact-key `GetItem` (n/a) | §3.4 — switch to `sharedaws.DynamoDBAPI` |
| `CreateIncident` | `repository.go:134` | `TransactWriteItems` (n/a) | §3.4 — switch to `sharedaws.DynamoDBAPI` |

Escalation-runtime has no primary-index scan or prefix queries today, so the
inventory focuses on the facade migration.

## Summary of methods that lose or cannot expose continuation

| Method | Runtime | Status today | Plan |
| --- | --- | --- | --- |
| `ListServices` | monitor-api | Unbounded (no cursor) | §4.4 page helper |
| `ListMonitors` | monitor-api | Unbounded | §4.4 page helper |
| `ListIncidents` | monitor-api | Unbounded | §4.4 page helper |
| `ListServiceIncidents` | monitor-api | Truncated (silently drops `LastEvaluatedKey`) | §4.4 page helper |
| `ListIncidentActivities` | monitor-api | Unbounded | §4.4 page helper |
| `serviceMonitorSummaries` | monitor-api | Transitive unbounded | §4.4 follow |
| `buildServiceStatusRecord` | monitor-api | Transitive unbounded | §4.4 follow |
| `GetServiceCardMetrics` | monitor-api | Bounded (Limit=20) but only one page returned | §6.2 regression |
| `ListMonitors` | check-runtime | Per-partition scan without cursor | §4.5 bounded helper |
| `ClaimExecutionWork` | check-runtime | Per-partition Query without cursor | §4.5 bounded helper |

The §4 helpers (`PageOf[Item]` with explicit `Limit`, `Cursor`, `Direction`,
plus `PageKey`) become the only legal way to issue a primary-index read from
§4.4 onward. Callers either supply the returned `PageKey` to fetch the next
page or accept bounded evidence with `Incomplete=true`.
