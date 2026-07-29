## Context

Manual-run records contain `ExpiresAt`, but repository persistence omits TTL. DynamoDB never removes keys and retention promise is broken.

## Goals / Non-Goals

**Goals:** Store absolute expiry in existing table TTL attribute. Preserve replay correctness during DynamoDB TTL deletion lag. Bound data retention and storage cost.

**Non-Goals:** New cleanup Lambda, scan/backfill, changing retention duration, or treating TTL deletion time as exact expiration.

## Decisions

- Persist `ExpiresAt.Unix()` into existing `TTL` attribute. DynamoDB TTL needs absolute epoch seconds, not retention duration.
- Keep logical expiry semantics explicit if record exists after expiry because TTL deletion is asynchronous.
- Reuse existing table configuration. TTL deletes cost no write capacity and prevents unbounded storage; no GSI, Lambda, schedule, or scan.

## Risks / Trade-offs

- TTL deletion can lag -> Request path must not depend on physical deletion at deadline.
- Existing records lack TTL -> They remain until separately removed; no destructive backfill in this fix.

## Migration Plan

Deploy repository change. Verify new records include epoch TTL matching configured retention. Monitor storage trend through normal table cost review. Rollback only affects newly written TTL fields, which are safe.
