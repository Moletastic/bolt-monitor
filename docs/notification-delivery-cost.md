# Notification Delivery Bounded Cost Model

This note covers the bounded incremental cost of the notification
delivery assurance work. It does not propose new integrations,
permanently-on resources, or human-receipt guarantees.

## Always-on Surface

No new always-on AWS resources beyond the dispatcher subscription
on the existing notification table stream.

- One DynamoDB Streams subscription on the existing primary table.
  Filter is INSERT-only on `EntityType=TransitionOutbox`. Filter
  policies are bounded and the filter is asserted in the SST test
  suite.

## Write Surface

- One canonical `TransitionOutbox` row per accepted notification-
  relevant incident transition. Existing prerequisite contract;
  this change consumes it.
- One `Delivery` row per incident × transition × step × channel.
  Conditionally created only after the immutable `EscalationPlan`
  for the step is persisted. Rows are bounded by `step × channel`
  count per incident transition.
- One `ReplayIdempotency` row per replay. TTL equals
  `ReplayIdempotencyRetention` (default 24h) — longer than the
  maximum replay dispatch + retry window. Expires automatically.
- Conditional `ReplayDispatch` outbox row per replay. The dispatcher
  reconciles pending items the same way as transition outbox rows.

## Sparse Pending Access Path

The dispatcher populates `DISPATCH_PENDING#<tenant>#<bucket>#<shard>`
as a durable recovery index. The reconciler reads only configured
recent buckets (`dispatchPendingBuckets`) and bounded pages
(`dispatchPendingShards`, `dispatchPendingPageLimit`). It does not
scan the primary table. Manual point repair uses a single
`TransitionOutbox` GetItem.

## Delivery Attempt Lease and Retries

- `DeliveryAttemptLease` is fenced; one in-flight claim per
  `deliveryId`. Lease expiry reverts the attempt to `ambiguous`.
- `NotificationQueueMaxReceiveCount` (8) bounds SQS redrives.
- `DeliveryAutomaticAttemptLimit` (5) bounds automatic retries.
  `timeout`, `transport`, `throttled`, `provider_5xx` are retryable;
  `provider_4xx` (other than 429), `invalid_config`,
  `unsupported_channel` are terminal.

## Scheduler One-Time Schedules

Each delayed step schedules one AWS Scheduler entry with
`ActionAfterCompletion=DELETE`. Schedules do not accumulate; they are
removed on completion. No annual cron rules or per-incident Lambda
permissions are added.

## Explicit Exclusions

This change does not add:

- Always-on polling or scanning services.
- Cross-region or multi-region traffic.
- New notification integrations or provider contracts.
- Human-receipt or acknowledgement tracking.
- Direct retry of Stream or Scheduler DLQ envelopes back to the
  notification queue.
- An on-call assignment or rotation platform.
