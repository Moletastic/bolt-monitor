## Context

The single-table DynamoDB design already treats `CheckRun` as high-volume append-only data and the Go model sets `TTL` to 30 days after creation. The SST table definition does not currently enable DynamoDB TTL, so those attributes do not cause deletion. `ExecutionWork` records are persisted under tenant partitions as execution coordination state and currently have no retention metadata.

This change is a FinOps retention improvement, not an archival feature. It favors deleting operational records after their useful hot window over copying them to S3.

## Goals / Non-Goals

**Goals:**
- Enable DynamoDB TTL in SST on the shared `TTL` attribute.
- Ensure new `CheckRun` records expire through native DynamoDB TTL after the existing 30-day window.
- Add TTL to `ExecutionWork` records so transient execution work does not accumulate indefinitely.
- Keep the retention implementation simple and native to DynamoDB.

**Non-Goals:**
- No S3 archival, Athena querying, Glacier lifecycle, or restore path.
- No retention changes for audit, incident, configuration, status, escalation, notification, or search index records.
- No active API behavior change beyond old expired records eventually disappearing from DynamoDB-backed reads.
- No table scans or custom archive/delete batch job.

## Decisions

### Use SST `ttl: 'TTL'` on the existing Dynamo table

SST 4.14.1 documents first-class Dynamo TTL support with `ttl: '<attributeName>'`. The existing `CheckRun` model already writes `TTL` as Unix epoch seconds, matching DynamoDB TTL requirements.

Alternative considered: Pulumi table transform for TTL. Rejected because first-class SST `ttl` is clearer, smaller, and documented for this SST version.

### Use one shared TTL attribute across eligible item families

The table should use the existing `TTL` attribute for all expiring items. DynamoDB ignores items without the TTL attribute, so persistent item families can remain unchanged.

Alternative considered: separate item-family-specific TTL attributes. Rejected because DynamoDB supports one TTL attribute per table and shared `TTL` is already present in run records.

### Delete `ExecutionWork` through TTL instead of adding a cleanup job

Execution work is operational coordination state, not business history. A TTL field keeps cleanup native and avoids new Lambdas, schedules, permissions, and failure modes.

Alternative considered: scheduled cleanup Lambda querying `RUN_REQUEST#` records. Rejected for first cut because it adds moving parts and risks tenant-partition scans.

### Keep archive out of scope

For this FinOps pass, deletion is more valuable than cold retention. Runs and execution work are high-volume operational records; audit and incident retention require separate product/compliance decisions.

Alternative considered: S3 archive before delete. Rejected for this change because it increases cost and operational surface before archive retrieval requirements exist.

## Risks / Trade-offs

- Expired items are not deleted exactly at the retention boundary → Accept DynamoDB TTL's eventual deletion model; APIs that require strict windows would need explicit timestamp filtering in a separate change.
- Enabling TTL may delete old `CheckRun` rows that already have expired `TTL` values → This is intended but should be called out during deployment.
- `ExecutionWork` TTL window chosen too short could remove useful debugging context → Use a conservative short operational window and keep recent failures visible long enough for investigation.
- Items without numeric `TTL` will not expire → Tests should assert new `CheckRun` and `ExecutionWork` records include numeric epoch-second TTLs.
- TTL deletes do not archive data → Product/compliance history for audit and incidents remains a separate future design.

## Migration Plan

1. Add `ttl: 'TTL'` to the SST Dynamo table definition.
2. Add `TTL` to `ExecutionWork` record shape and set it when creating/updating work records.
3. Keep existing `CheckRun` TTL calculation unchanged unless tests reveal drift.
4. Deploy to staging and verify table TTL is enabled on `TTL`.
5. Observe that expired items are removed asynchronously by DynamoDB.

Rollback: remove the SST TTL setting to stop future TTL deletions. DynamoDB TTL deletion that already occurred cannot be undone.

## Open Questions

- Exact `ExecutionWork` retention window: likely 7 days, but implementation should confirm whether manual-run debugging needs longer.
