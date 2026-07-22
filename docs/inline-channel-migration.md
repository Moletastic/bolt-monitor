# Legacy Inline-Channel Migration

Escalation-policy reads never repair legacy inline channel data. Policies created
before channel IDs were introduced retain their inline channel payload until an
operator runs the explicit migration.

Run against the intended target credentials:

```bash
TABLE_NAME=<app-table> MIGRATE_INLINE_CHANNELS=yes \
  go run -tags inline_channel_migration ./services/monitor-api
```

Set `TENANT_ID` only when migrating a tenant other than `DEFAULT`.

The command creates deterministic channel IDs from policy and step identity,
replaces inline route payloads with those IDs, and can be run again safely.
It reports scanned and migrated policy counts. Stop on an error, correct the
underlying DynamoDB or credential problem, then rerun; policies already
converted remain unchanged.
