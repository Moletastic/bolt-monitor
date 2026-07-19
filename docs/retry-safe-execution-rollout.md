# Retry-Safe Execution Rollout

## Deploy

1. Pause recurring scheduling and reject manual run commands.
2. Wait for execution queue depth to reach zero and active workers to finish.
3. Deploy the canonical transition outbox and sparse `DISPATCH_PENDING` index.
4. Verify scheduler polling dispatch is enabled before allowing transition-producing checks.
5. Deploy scheduler, worker, and escalation runtime together. The worker does not send directly to the notification queue.
6. Smoke-test one manual idempotency replay, one recurring run, one duplicate execution delivery, one publication-marker recovery, and one transition dispatch.
7. Re-enable manual runs and recurring scheduling.

The dispatcher is a bounded scheduler poll over four current tenant/hour shards. It intentionally does not use DynamoDB Streams, avoiding stream shard-hour cost for low-volume deployments. Older pending buckets are retained and require explicit recovery after a dispatcher outage; do not widen polling blindly because idle queries are a direct DynamoDB cost.

## Rollback

1. Pause manual and recurring producers.
2. Drain worker queue and wait for active leases to expire or complete.
3. Retain pending transition outbox and dispatch-pending records; do not delete or bypass them.
4. Restore previous runtime code only after drain. Existing retry-safe work, idempotency, raw run, and outbox items expire or remain durable under their configured policies.
5. Do not create a replacement queue, table, GSI, or backfill missed schedules during rollback.

Legacy work records retain their existing TTL. The deployment drain avoids indefinite compatibility code for old execution envelopes.
